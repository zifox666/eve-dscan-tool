from typing import List, Dict, Any
import logging
import asyncio
from httpx import HTTPError

from api.client import get_http_client
from config import settings
from utils.character_cache import character_cache

logger = logging.getLogger(__name__)


async def make_api_request(method, url, params=None, json=None, max_retries=3, retry_delay=1):
    """通用API请求函数，支持重试机制"""
    client = get_http_client()
    retries = 0

    while retries <= max_retries:
        try:
            if method == "post":
                response = await client.post(url, params=params, json=json)
            else:
                response = await client.get(url, params=params)

            response.raise_for_status()
            return response.json()
        except HTTPError as e:
            retries += 1
            if retries > max_retries:
                logger.error(f"请求失败 ({url}): {str(e)}")
                return None

            logger.warning(f"请求失败，正在重试 {retries}/{max_retries}: {str(e)}")
            await asyncio.sleep(retry_delay * retries)

    return None


async def get_character_ids(character_names: List[str]) -> Dict[str, int]:
    """根据角色名称获取角色ID"""
    result = {}

    unique_names = list(set(character_names))

    for name in unique_names:
        cached_id = character_cache.get_character_id(name)
        if cached_id:
            result[name] = cached_id

    uncached_names = character_cache.filter_cached_names(unique_names)

    if uncached_names:
        for i in range(0, len(uncached_names), 500):
            batch = uncached_names[i:i + 500]

            data = await make_api_request(
                "post",
                f"{settings.ESI_BASE_URL}/universe/ids/",
                params={
                    "datasource": settings.ESI_DATASOURCE,
                    "language": settings.ESI_LANGUAGE
                },
                json=batch
            )

            if data and "characters" in data:
                for character in data["characters"]:
                    char_name = character["name"]
                    char_id = character["id"]
                    result[char_name] = char_id
                    character_cache.set_character_id(char_name, char_id)

    return result


async def get_character_affiliations(character_ids: List[int]) -> List[Dict[str, Any]]:
    """获取角色所属的公司和联盟信息"""
    all_affiliations = []

    for char_id in character_ids:
        cached_info = character_cache.get_character_affiliation(char_id)
        if cached_info:
            all_affiliations.append(cached_info)

    uncached_ids = character_cache.filter_cached_ids(character_ids)

    if uncached_ids:
        for i in range(0, len(uncached_ids), settings.MAX_CHARACTERS_PER_BATCH):
            batch = uncached_ids[i:i + settings.MAX_CHARACTERS_PER_BATCH]

            affiliations = await make_api_request(
                "post",
                f"{settings.ESI_BASE_URL}/characters/affiliation/",
                params={"datasource": settings.ESI_DATASOURCE},
                json=batch
            )

            if affiliations:
                for affiliation in affiliations:
                    all_affiliations.append(affiliation)
                    character_cache.set_character_affiliation(affiliation["character_id"], affiliation)

    return all_affiliations


async def get_names_for_ids(ids: List[int]) -> Dict[int, Dict[str, str]]:
    """获取ID对应的名称和类别"""
    result = {}

    for entity_id in ids:
        cached_data = character_cache.get_name_for_id(entity_id)
        if cached_data:
            result[entity_id] = cached_data

    uncached_ids = character_cache.filter_cached_entity_ids(ids)

    if uncached_ids:
        for i in range(0, len(uncached_ids), settings.MAX_CHARACTERS_PER_BATCH):
            batch = uncached_ids[i:i + settings.MAX_CHARACTERS_PER_BATCH]

            names = await make_api_request(
                "post",
                f"{settings.ESI_BASE_URL}/universe/names/",
                params={"datasource": settings.ESI_DATASOURCE},
                json=batch
            )

            if names:
                for item in names:
                    entity_id = item["id"]
                    data = {
                        "name": item["name"],
                        "category": item["category"]
                    }
                    result[entity_id] = data
                    character_cache.set_name_for_id(entity_id, data)

    return result
