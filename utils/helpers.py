import random
import string
import re
from typing import List, Dict, Any, Optional, Tuple
from datetime import datetime, timedelta

# 生成随机短链接ID
def generate_short_id(length: int = 10) -> str:
    """生成指定长度的随机字符串作为短链接ID"""
    chars = string.ascii_letters + string.digits
    return ''.join(random.choice(chars) for _ in range(length))

# 格式化时间差
def format_time_ago(dt: datetime) -> str:
    """将时间格式化为'xx分钟前'的形式"""
    now = datetime.utcnow()
    diff = now - dt
    
    if diff < timedelta(minutes=1):
        return "刚刚"
    elif diff < timedelta(hours=1):
        minutes = diff.seconds // 60
        return f"{minutes} 分钟前"
    elif diff < timedelta(days=1):
        hours = diff.seconds // 3600
        return f"{hours} 小时前"
    else:
        days = diff.days
        return f"{days} 天前"

# 检测DScan类型
def detect_dscan_type(dscan_text: str) -> str:
    """检测DScan的类型（local或ship）"""
    lines = [line.strip() for line in dscan_text.strip().split('\n') if line.strip()]
    
    # 如果每行都只有一个单词或者包含空格的名称，可能是local dscan
    if all(len(line.split('\t')) == 1 for line in lines):
        return "local"
    else:
        return "ship"
        
# 提取星系信息
def extract_system_info(dscan_data: str) -> dict:
    """从DScan数据中提取星系信息"""
    system_info = {
        "system_name": None,
        "region_name": None
    }
    
    # 尝试从数据中匹配星系信息
    # 常见格式: "4-HWWF - SG.CN LifeStyle Market"
    system_pattern = r'([A-Z0-9\-]+)\s*-\s*'
    match = re.search(system_pattern, dscan_data)
    if match:
        system_info["system_name"] = match.group(1).strip()
    
    return system_info

# 解析Local DScan数据
def parse_local_dscan(dscan_text: str) -> List[str]:
    """解析Local DScan数据，返回角色名称列表"""
    lines = [line.strip() for line in dscan_text.strip().split('\n') if line.strip()]
    return lines

# 组织Local DScan结果数据
def organize_local_dscan_data(character_names: List[str], affiliations: List[Dict[str, Any]], names: Dict[int, Dict[str, str]]) -> Dict[str, Any]:
    """组织Local DScan的结果数据，用于前端展示和存储"""
    # 初始化结果数据结构
    result = {
        "alliances": {},  # 联盟信息
        "corporations": {},  # 军团信息
        "characters": {},  # 角色信息
        "stats": {  # 统计信息
            "alliance_count": 0,
            "corporation_count": 0,
            "character_count": 0,
        }
    }
    
    # 处理角色和所属信息
    for affiliation in affiliations:
        character_id = affiliation["character_id"]
        corporation_id = affiliation["corporation_id"]
        alliance_id = affiliation.get("alliance_id")
        
        # 添加角色信息
        if character_id in names:
            character_name = names[character_id]["name"]
            # 移除角色名称中的缩写标签
            if '[' in character_name and ']' in character_name:
                character_name = character_name.split('[')[0].strip()
                
            result["characters"][character_id] = {
                "id": character_id,
                "name": character_name,
                "corporation_id": corporation_id,
                "alliance_id": alliance_id
            }
        
        # 添加军团信息
        if corporation_id not in result["corporations"] and corporation_id in names:
            corp_name = names[corporation_id]["name"]
            # 提取公司缩写（ticker）
            corp_ticker = ""
            if '[' in corp_name and ']' in corp_name:
                parts = corp_name.split('[', 1)
                corp_name = parts[0].strip()
                if len(parts) > 1 and ']' in parts[1]:
                    corp_ticker = parts[1].split(']')[0].strip()
                    
            result["corporations"][corporation_id] = {
                "id": corporation_id,
                "name": corp_name,
                "ticker": corp_ticker,
                "alliance_id": alliance_id,
                "character_count": 0
            }
        
        # 增加军团的角色计数
        if corporation_id in result["corporations"]:
            result["corporations"][corporation_id]["character_count"] += 1
        
        # 添加联盟信息
        if alliance_id and alliance_id not in result["alliances"] and alliance_id in names:
            alliance_name = names[alliance_id]["name"]
            # 提取联盟缩写（ticker）
            alliance_ticker = ""
            if '[' in alliance_name and ']' in alliance_name:
                parts = alliance_name.split('[', 1)
                alliance_name = parts[0].strip()
                if len(parts) > 1 and ']' in parts[1]:
                    alliance_ticker = parts[1].split(']')[0].strip()
                    
            result["alliances"][alliance_id] = {
                "id": alliance_id,
                "name": alliance_name,
                "ticker": alliance_ticker,
                "corporation_count": 0,
                "character_count": 0
            }
        
        # 增加联盟的角色计数
        if alliance_id and alliance_id in result["alliances"]:
            result["alliances"][alliance_id]["character_count"] += 1
    
    # 确保每个公司都有正确的联盟关联
    for corp_id, corp_info in result["corporations"].items():
        alliance_id = corp_info.get("alliance_id")
        if alliance_id and alliance_id in result["alliances"]:
            result["alliances"][alliance_id]["corporation_count"] += 1
    
    # 更新统计信息
    result["stats"]["alliance_count"] = len(result["alliances"])
    result["stats"]["corporation_count"] = len(result["corporations"])
    result["stats"]["character_count"] = len(result["characters"])
    
    return result