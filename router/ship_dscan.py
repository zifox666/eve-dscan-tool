from typing import List, Dict, Any

from fastapi import APIRouter, Request, Form, Depends
from fastapi.responses import HTMLResponse, RedirectResponse, JSONResponse
from fastapi.templating import Jinja2Templates
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from config import settings
from db.database import get_db
from db.models.dscan import ShipDScan
from db.sde import eve_db
from utils.cache import dscan_cache
from utils.helpers import generate_short_id, format_time_ago
from utils.sqlite_helper import SQLiteHelper
from utils.system_list import system_list_zh

import re, math

router = APIRouter()

templates = Jinja2Templates(directory=settings.TEMPLATES_DIR)


def parse_ship_dscan(dscan_text: str) -> tuple[list[dict[str, int | str | None]], dict]:
    """解析舰船扫描数据，返回舰船信息列表"""
    lines = [line.strip() for line in dscan_text.strip().split('\n') if line.strip()]
    result = []

    for line in lines:
        parts = line.split('\t')
        if len(parts) >= 2:
            item = {}

            try:
                item['type_id'] = int(parts[0].strip())
            except ValueError:
                continue

            item['name'] = parts[1].strip()

            if len(parts) >= 3:
                item['type_name'] = parts[2].strip()
            else:
                item['type_name'] = None

            if len(parts) >= 4:
                item['distance'] = parts[3].strip()
            else:
                item['distance'] = None

            result.append(item)

    return result


def get_ship_info_from_db(ship_items: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
    """从SQLite数据库获取舰船信息"""
    sqlite_helper = SQLiteHelper(settings.SQLITE_DB_PATH)
    if not sqlite_helper.connect():
        return ship_items

    try:
        type_ids = [item['type_id'] for item in ship_items]

        type_infos = sqlite_helper.get_type_infos(type_ids)

        for item in ship_items:
            type_id = item['type_id']
            if type_id in type_infos:
                info = type_infos[type_id]
                if info.get('zh_name'):
                    item['type_name'] = info['zh_name']
                elif not item['type_name'] and info.get('name'):
                    item['type_name'] = info['name']

                item['group_name'] = info.get('group_name')
                item['category_name'] = info.get('category_name')
    finally:
        sqlite_helper.close()

    return ship_items


def organize_ship_dscan_data(ship_items: List[Dict[str, Any]], language: str = "zh", filter_distance: bool = False) -> Dict[str, Any]:
    """组织舰船扫描数据，按类型和分组分类，支持多语言"""
    # 旗舰类型组ID列表
    capital_group_ids = [30, 485, 547, 659, 883, 1538, 4594]

    # 建筑类型分类ID列表
    structure_category_ids = [41, 46, 65, 66]

    result = {
        "ship_types": {},
        "capital_types": {},
        "structure_types": {},
        "misc_types": {},
        "capital_group_ids": capital_group_ids,
        "stats": {
            "total_count": len(ship_items),
            "ship_count": 0,
            "capital_count": 0,
            "structure_count": 0,
            "misc_count": 0
        },
        "system_info": {
            "name": None,
            "id": None,
            "region": None,
            "constellation": None
        },
    }

    system_info_candidates = {}

    

# def extract_system_info(dscan_data: str) -> dict:
#     """从DScan数据中提取星系信息"""
#     system_info = {
#         "system_name": None,
#         "region_name": None
#     }

#     # 常见格式: "4-HWWF - SG.CN LifeStyle Market"
#     match = re.search(system_pattern, dscan_data)
#     if match:
#         system_info["system_name"] = match.group(1).strip()

#     return system_info

    system_pattern = r'^(.+?)( [XIV]+)? - '

    for item in ship_items:
        type_id = item.get('type_id')
        if not type_id:
            continue

        type_info = eve_db.get_type_info(type_id, language)
        if not type_info:
            continue

        category_id = type_info.get('categoryID')
        match = re.search(system_pattern, item.get('name'))
        if (category_id != 6) and match:
            if match.group(1).strip() not in system_info_candidates:
                system_info_candidates[match.group(1).strip()] = 1
            else:
                system_info_candidates[match.group(1).strip()] += 1
        
        if filter_distance and not (item.get('distance') and item.get('distance') != '-'):
            continue

        group_id = type_info.get('groupID')
        group_name = type_info.get('group_name', '未知分组')
        type_name = type_info.get('name', 'Unknown')

        if category_id == 6:  # 舰船分类ID
            # 普通舰船处理
            if group_name not in result["ship_types"]:
                result["ship_types"][group_name] = {}

            if type_name in result["ship_types"][group_name]:
                result["ship_types"][group_name][type_name] += 1
            else:
                result["ship_types"][group_name][type_name] = 1

            result["stats"]["ship_count"] += 1

            # 旗舰处理
            if group_id in capital_group_ids:
                result["stats"]["capital_count"] += 1

                if group_name not in result["capital_types"]:
                    result["capital_types"][group_name] = {}

                if type_name in result["capital_types"][group_name]:
                    result["capital_types"][group_name][type_name] += 1
                else:
                    result["capital_types"][group_name][type_name] = 1


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

    for category in ["ship_types", "capital_types", "structure_types", "misc_types"]:
        sorted_groups = sorted(
            result[category].items(),
            key=lambda x: sum(x[1].values()),
            reverse=True
        )

        # 组内按数量排序类型
        result[category] = {
            group: [{"name": type_name, "count": count}
                    for type_name, count in sorted(types.items(), key=lambda x: x[1], reverse=True)]
            for group, types in sorted_groups
        }
    system_name = None
    if system_info_candidates:
        system_name, _ = max(system_info_candidates.items(), key=lambda item: item[1])

    if system_name is not None:
        result["system_info"]["name"] = system_name

        if language == "zh" and system_name in system_list_zh:
            result["system_info"]["id"] = system_list_zh[system_name]
        
        if language == "en":
            system_id = eve_db.get_system_id_by_en_name(system_name)
            if system_id:
                result["system_info"]["id"] = system_id

        if result["system_info"]["id"]:
            info = eve_db.get_system_info_by_id(result["system_info"]["id"])
            result["system_info"]["region"] = info.get("regionID")
            result["system_info"]["regionName"] = info.get("regionName")
            result["system_info"]["constellation"] = info.get("constellationID")
            result["system_info"]["constellationName"] = info.get("constellationName")
            result["system_info"]["security"] = math.floor(info.get("security") * 100) / 100
        
    return result


@router.post("/process", response_class=HTMLResponse)
async def process_ship_dscan(
        request: Request,
        data: str = Form(...),
        filter_distance: bool = Form(False),
        db: AsyncSession = Depends(get_db)
):
    """处理Ship DScan数据，同时生成中英文版本"""
    ship_items = parse_ship_dscan(data)

    if not ship_items:
        return templates.TemplateResponse(
            "404.html",
            {"request": request, "message": "无法解析舰船扫描数据，请检查格式是否正确"}
        )


    short_id = generate_short_id(settings.SHORT_LINK_LENGTH)

    client_ip = request.client.host

    result_zh = organize_ship_dscan_data(ship_items, "zh", filter_distance)
    result_en = organize_ship_dscan_data(ship_items, "en", filter_distance)

    for result in [result_zh, result_en]:
        result["filter_distance"] = filter_distance

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

    if "application/json" in request.headers.get("accept", ""):
        return JSONResponse(
            status_code=201,
            content={
                "code": 201,
                "msg": "成功",
                "data": {
                    "short_id": new_dscan.short_id,
                    "view_url": f"/v/{new_dscan.short_id}",
                }
            }
        )

    return RedirectResponse(url=f"/v/{short_id}", status_code=303)


@router.get("/{short_id}")
async def view_ship_dscan(
        request: Request,
        short_id: str,
        db: AsyncSession = Depends(get_db)
):
    """查看保存的Ship DScan数据，支持语言切换，根据Accept头决定返回HTML或JSON"""
    lang = request.cookies.get("lang", "zh")
    if lang not in ["zh", "en"]:
        lang = "zh"

    cache_key = f"ship_dscan:{lang}:{short_id}"
    cached_data = dscan_cache.get(cache_key)

    query = select(ShipDScan).where(ShipDScan.short_id == short_id)
    result = await db.execute(query)
    dscan = result.scalar_one_or_none()

    if not dscan:
        if "application/json" in request.headers.get("accept", ""):
            return JSONResponse(
                status_code=404,
                content={
                    "code": 404,
                    "msg": "找不到指定的舰船扫描数据",
                    "data": {}
                }
            )
        return templates.TemplateResponse(
            "404.html",
            {"request": request, "message": "找不到指定的舰船扫描数据"}
        )

    dscan.view_count += 1
    await db.commit()

    if cached_data:
        dscan_data = cached_data
        dscan_data["view_count"] = dscan.view_count
    else:
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

        dscan_cache.set(cache_key, dscan_data)

    if "application/json" in request.headers.get("accept", ""):
        return {
            "code": 200,
            "msg": "成功",
            "data": dscan_data
        }

    return templates.TemplateResponse(
        "ship_dscan.html.jinja2",
        {
            "request": request,
            "title": "舰船扫描分析",
            "dscan": dscan_data,
            "base_url": f"{request.base_url}v/{short_id}",
            "current_lang": lang
        }
    )
