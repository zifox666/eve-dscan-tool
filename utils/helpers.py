import random
import string
from datetime import datetime, timedelta, timezone
from typing import List, Dict, Any


def generate_short_id(length: int = 10) -> str:
    """生成指定长度的随机字符串作为短链接ID"""
    chars = string.ascii_letters + string.digits
    return ''.join(random.choice(chars) for _ in range(length))


def format_time_ago(dt: datetime) -> str:
    """将时间格式化为'xx分钟前'的形式"""
    now = datetime.utcnow()
    try:
        diff = now - dt

        if diff < timedelta(minutes=1):
            return "Recently"
        elif diff < timedelta(hours=1):
            minutes = diff.seconds // 60
            return f"{minutes} min ago"
        elif diff < timedelta(days=1):
            hours = diff.seconds // 3600
            return f"{hours} hours ago"
        else:
            days = diff.days
            return f"{days} days ago"
    except:
        return "Unknown time"


def detect_dscan_type(dscan_text: str) -> str:
    """检测DScan的类型（local或ship）"""
    lines = [line.strip() for line in dscan_text.strip().split('\n') if line.strip()]

    if all(len(line.split('\t')) == 1 for line in lines):
        return "local"
    else:
        return "ship"

def parse_local_dscan(dscan_text: str) -> List[str]:
    """解析Local DScan数据，返回角色名称列表"""
    lines = [line.strip() for line in dscan_text.strip().split('\n') if line.strip()]
    return lines


def organize_local_dscan_data(
        character_names: List[str],
        affiliations: List[Dict[str, Any]],
        names: Dict[int, Dict[str, str]]
) -> Dict[str, Any]:
    """
    组织Local DScan的结果数据，用于前端展示和存储

    :param character_names: 角色名称列表(废弃)
    :param affiliations: 角色的联盟和军团信息列表
    :param names: 角色、军团和联盟的名称字典

    :return: 组织好的数据字典
    """
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

    for affiliation in affiliations:
        character_id = affiliation["character_id"]
        corporation_id = affiliation["corporation_id"]
        alliance_id = affiliation.get("alliance_id")

        if character_id in names:
            character_name = names[character_id]["name"]
            if '[' in character_name and ']' in character_name:
                character_name = character_name.split('[')[0].strip()

            result["characters"][character_id] = {
                "id": character_id,
                "name": character_name,
                "corporation_id": corporation_id,
                "alliance_id": alliance_id
            }

        if corporation_id not in result["corporations"] and corporation_id in names:
            corp_name = names[corporation_id]["name"]
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

        if corporation_id in result["corporations"]:
            result["corporations"][corporation_id]["character_count"] += 1

        if alliance_id and alliance_id not in result["alliances"] and alliance_id in names:
            alliance_name = names[alliance_id]["name"]
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

        if alliance_id and alliance_id in result["alliances"]:
            result["alliances"][alliance_id]["character_count"] += 1

    for corp_id, corp_info in result["corporations"].items():
        alliance_id = corp_info.get("alliance_id")
        if alliance_id and alliance_id in result["alliances"]:
            result["alliances"][alliance_id]["corporation_count"] += 1

    result["stats"]["alliance_count"] = len(result["alliances"])
    result["stats"]["corporation_count"] = len(result["corporations"])
    result["stats"]["character_count"] = len(result["characters"])

    return result
