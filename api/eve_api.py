from typing import List, Dict, Any, Optional
import json
from api.client import get_http_client
from config import settings

# 获取角色ID
async def get_character_ids(character_names: List[str]) -> Dict[str, int]:
    """根据角色名称获取角色ID"""
    # 分批处理，每批最多200个角色
    result = {}
    for i in range(0, len(character_names), settings.MAX_CHARACTERS_PER_BATCH):
        batch = character_names[i:i + settings.MAX_CHARACTERS_PER_BATCH]
        
        # 调用ESI API
        client = get_http_client()
        response = await client.post(
            f"{settings.ESI_BASE_URL}/universe/ids/",
            params={
                "datasource": settings.ESI_DATASOURCE,
                "language": settings.ESI_LANGUAGE
            },
            json=batch
        )
        response.raise_for_status()
        data = response.json()
        
        # 处理返回的角色数据
        if "characters" in data:
            for character in data["characters"]:
                result[character["name"]] = character["id"]
    
    return result

# 获取角色所属信息
async def get_character_affiliations(character_ids: List[int]) -> List[Dict[str, Any]]:
    """获取角色所属的公司和联盟信息"""
    # 分批处理，每批最多200个角色
    all_affiliations = []
    for i in range(0, len(character_ids), settings.MAX_CHARACTERS_PER_BATCH):
        batch = character_ids[i:i + settings.MAX_CHARACTERS_PER_BATCH]
        
        # 调用ESI API
        client = get_http_client()
        response = await client.post(
            f"{settings.ESI_BASE_URL}/characters/affiliation/",
            params={"datasource": settings.ESI_DATASOURCE},
            json=batch
        )
        response.raise_for_status()
        affiliations = response.json()
        all_affiliations.extend(affiliations)
    
    return all_affiliations

# 获取ID对应的名称
async def get_names_for_ids(ids: List[int]) -> Dict[int, Dict[str, str]]:
    """获取ID对应的名称和类别"""
    # 分批处理，每批最多200个ID
    result = {}
    for i in range(0, len(ids), settings.MAX_CHARACTERS_PER_BATCH):
        batch = ids[i:i + settings.MAX_CHARACTERS_PER_BATCH]
        
        # 调用ESI API
        client = get_http_client()
        response = await client.post(
            f"{settings.ESI_BASE_URL}/universe/names/",
            params={"datasource": settings.ESI_DATASOURCE},
            json=batch
        )
        response.raise_for_status()
        names = response.json()
        
        # 处理返回的名称数据
        for item in names:
            result[item["id"]] = {
                "name": item["name"],
                "category": item["category"]
            }
    
    return result