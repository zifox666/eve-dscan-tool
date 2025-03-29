from fastapi import APIRouter, Request, Form, Depends, Query
from fastapi.responses import HTMLResponse, JSONResponse, RedirectResponse
from fastapi.templating import Jinja2Templates
from redis.asyncio import Redis
from sqlalchemy.ext.asyncio import AsyncSession
from sqlalchemy import select
import json
import re
from typing import List, Dict, Any, Optional, Tuple

from config import settings
from db.database import get_db
from db.models.dscan import ShipDScan
from db.sde import eve_db
from utils.cache import dscan_cache
from utils.helpers import generate_short_id, format_time_ago
from utils.sqlite_helper import SQLiteHelper
from utils.time import DateTimeEncoder

# 创建路由
router = APIRouter()

# 设置模板
templates = Jinja2Templates(directory=settings.TEMPLATES_DIR)

# 解析舰船扫描数据
def parse_ship_dscan(dscan_text: str) -> List[Dict[str, Any]]:
    """解析舰船扫描数据，返回舰船信息列表"""
    lines = [line.strip() for line in dscan_text.strip().split('\n') if line.strip()]
    result = []
    
    # 提取星系信息
    from utils.helpers import extract_system_info
    system_info = extract_system_info(dscan_text)
    
    for line in lines:
        # 尝试解析不同格式的DScan数据
        parts = line.split('\t')
        if len(parts) >= 2:
            item = {}
            # 解析类型ID
            try:
                item['type_id'] = int(parts[0].strip())
            except ValueError:
                continue  # 跳过无法解析ID的行
            
            # 解析名称
            item['name'] = parts[1].strip()
            
            # 解析类型名称（如果有）
            if len(parts) >= 3:
                item['type_name'] = parts[2].strip()
            else:
                item['type_name'] = None
            
            # 解析距离（如果有）
            if len(parts) >= 4:
                item['distance'] = parts[3].strip()
            else:
                item['distance'] = None
            
            result.append(item)
    
    return result, system_info

# 从SQLite数据库获取舰船信息
def get_ship_info_from_db(ship_items: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """从SQLite数据库获取舰船信息"""
    # 连接SQLite数据库
    sqlite_helper = SQLiteHelper(settings.SQLITE_DB_PATH)
    if not sqlite_helper.connect():
        # 如果连接失败，返回原始数据
        return ship_items
    
    try:
        # 获取所有类型ID
        type_ids = [item['type_id'] for item in ship_items]
        
        # 批量查询类型信息
        type_infos = sqlite_helper.get_type_infos(type_ids)
        
        # 更新舰船信息
        for item in ship_items:
            type_id = item['type_id']
            if type_id in type_infos:
                info = type_infos[type_id]
                # 使用中文名称（如果有）
                if info.get('zh_name'):
                    item['type_name'] = info['zh_name']
                elif not item['type_name'] and info.get('name'):
                    item['type_name'] = info['name']
                
                # 添加分组和分类信息
                item['group_name'] = info.get('group_name')
                item['category_name'] = info.get('category_name')
    finally:
        # 关闭数据库连接
        sqlite_helper.close()
    
    return ship_items

# 组织舰船扫描数据
def organize_ship_dscan_data(ship_items: List[Dict[str, Any]], language: str = "zh") -> Dict[str, Any]:
    """组织舰船扫描数据，按类型和分组分类，支持多语言"""
    # 旗舰类型组ID列表
    capital_group_ids = [30, 485, 547, 659, 883, 1538, 4594]

    # 建筑类型分类ID列表
    structure_category_ids = [41, 46, 65, 66]

    # 初始化结果数据结构
    result = {
        "ship_types": {},
        "capital_types": {},
        "structure_types": {},
        "misc_types": {},
        "stats": {
            "total_count": len(ship_items),
            "ship_count": 0,
            "capital_count": 0,
            "structure_count": 0,
            "misc_count": 0
        }
    }

    # 分类舰船数据
    for item in ship_items:
        type_id = item.get('type_id')
        if not type_id:
            continue

        # 从SDE数据库获取对应语言的信息
        type_info = eve_db.get_type_info(type_id, language)
        if not type_info:
            continue

        group_id = type_info.get('groupID')
        category_id = type_info.get('categoryID')
        group_name = type_info.get('group_name', '未知分组')
        type_name = type_info.get('name', 'Unknown')

        # 判断分类
        if category_id == 6:  # 舰船分类ID
            if group_id in capital_group_ids:
                # 旗舰处理
                if group_name not in result["capital_types"]:
                    result["capital_types"][group_name] = {}

                if type_name in result["capital_types"][group_name]:
                    result["capital_types"][group_name][type_name] += 1
                else:
                    result["capital_types"][group_name][type_name] = 1

                result["stats"]["capital_count"] += 1
            else:
                # 普通舰船处理
                if group_name not in result["ship_types"]:
                    result["ship_types"][group_name] = {}

                if type_name in result["ship_types"][group_name]:
                    result["ship_types"][group_name][type_name] += 1
                else:
                    result["ship_types"][group_name][type_name] = 1

                result["stats"]["ship_count"] += 1
        elif category_id in structure_category_ids:
            # 建筑处理
            if group_name not in result["structure_types"]:
                result["structure_types"][group_name] = {}

            if type_name in result["structure_types"][group_name]:
                result["structure_types"][group_name][type_name] += 1
            else:
                result["structure_types"][group_name][type_name] = 1

            result["stats"]["structure_count"] += 1
        else:
            # 其他物品处理
            if group_name not in result["misc_types"]:
                result["misc_types"][group_name] = {}

            if type_name in result["misc_types"][group_name]:
                result["misc_types"][group_name][type_name] += 1
            else:
                result["misc_types"][group_name][type_name] = 1

            result["stats"]["misc_count"] += 1

    # 将嵌套字典转换为便于前端使用的格式，并对组进行排序
    for category in ["ship_types", "capital_types", "structure_types", "misc_types"]:
        # 按组内总数量排序组
        sorted_groups = sorted(
            result[category].items(),
            key=lambda x: sum(x[1].values()),
            reverse=True
        )

        # 重构分类数据，按组内总数量排序组，每个组内按数量排序类型
        result[category] = {
            group: [{"name": type_name, "count": count}
                    for type_name, count in sorted(types.items(), key=lambda x: x[1], reverse=True)]
            for group, types in sorted_groups
        }

    return result


@router.post("/process", response_class=HTMLResponse)
async def process_ship_dscan(
        request: Request,
        data: str = Form(...),
        filter_distance: bool = Form(False),
        db: AsyncSession = Depends(get_db)
):
    """处理Ship DScan数据，同时生成中英文版本"""
    # 解析DScan数据
    ship_items, system_info = parse_ship_dscan(data)

    if not ship_items:
        return templates.TemplateResponse(
            "404.html",
            {"request": request, "message": "无法解析舰船扫描数据，请检查格式是否正确"}
        )

    if filter_distance:
        ship_items = [item for item in ship_items if item.get('distance') and item.get('distance') != '-']

    # 生成短链接ID
    short_id = generate_short_id(settings.SHORT_LINK_LENGTH)

    # 获取客户端IP
    client_ip = request.client.host

    # 分别生成中英文数据
    result_zh = organize_ship_dscan_data(ship_items, "zh")
    result_en = organize_ship_dscan_data(ship_items, "en")

    # 添加过滤和星系信息
    for result in [result_zh, result_en]:
        result["filter_distance"] = filter_distance
        result["system_info"] = system_info

    # 保存到数据库（存储两种语言的数据）
    processed_data = {
        "zh": result_zh,
        "en": result_en
    }

    new_dscan = ShipDScan(
        short_id=short_id,
        raw_data=data,
        processed_data=processed_data,
        client_ip=client_ip
    )
    db.add(new_dscan)
    await db.commit()
    await db.refresh(new_dscan)

    # 缓存处理后的数据到Redis（分别缓存中英文数据）
    for lang in ["zh", "en"]:
        cache_key = f"ship_dscan:{lang}:{short_id}"
        dscan_data = {
            "id": new_dscan.id,
            "short_id": new_dscan.short_id,
            "processed_data": processed_data[lang],
            "view_count": new_dscan.view_count,
            "created_at": new_dscan.created_at.isoformat(),
            "time_ago": format_time_ago(new_dscan.created_at)
        }
        dscan_cache.set(cache_key, dscan_data)

    # 重定向到查看页面
    return RedirectResponse(url=f"/v/{short_id}", status_code=303)


@router.get("/{short_id}", response_class=HTMLResponse)
async def view_ship_dscan(
        request: Request,
        short_id: str,
        db: AsyncSession = Depends(get_db)
):
    """查看保存的Ship DScan数据，支持语言切换"""
    # 获取用户语言偏好，默认为中文
    lang = request.cookies.get("lang", "zh")
    if lang not in ["zh", "en"]:
        lang = "zh"  # 默认为中文

    # 缓存键
    cache_key = f"ship_dscan:{lang}:{short_id}"
    cached_data = dscan_cache.get(cache_key)

    # 从数据库查询记录
    query = select(ShipDScan).where(ShipDScan.short_id == short_id)
    result = await db.execute(query)
    dscan = result.scalar_one_or_none()

    if not dscan:
        return templates.TemplateResponse(
            "404.html",
            {"request": request, "message": "找不到指定的舰船扫描数据"}
        )

    # 更新访问次数
    dscan.view_count += 1
    await db.commit()

    if cached_data:
        dscan_data = cached_data
        dscan_data["view_count"] = dscan.view_count
    else:
        # 缓存未命中，从数据库获取并构建相应语言的数据
        processed_data = dscan.processed_data.get(lang, dscan.processed_data.get("zh", {}))

        dscan_data = {
            "id": dscan.id,
            "short_id": dscan.short_id,
            "processed_data": processed_data,
            "view_count": dscan.view_count,
            "created_at": dscan.created_at.isoformat(),
            "time_ago": format_time_ago(dscan.created_at),
            "lang": lang
        }

        # 更新缓存
        dscan_cache.set(cache_key, dscan_data)

    return templates.TemplateResponse(
        "ship_dscan.html",
        {
            "request": request,
            "title": "舰船扫描分析",
            "dscan": dscan_data,
            "base_url": f"{request.base_url}v/{short_id}",
            "current_lang": lang
        }
    )