from datetime import datetime, timedelta
from typing import Dict, Any, Optional, List

from config import settings


class CharacterCache:
    def __init__(self):
        self.name_to_id: Dict[str, Dict[str, Any]] = {}
        self.id_to_info: Dict[int, Dict[str, Any]] = {}
        self.id_to_name: Dict[int, Dict[str, Any]] = {}
        self.ttl = timedelta(days=settings.TTL_CHARACTER_CACHE_DAYS)

    def get_character_id(self, name: str) -> Optional[int]:
        """根据角色名获取ID"""
        if name in self.name_to_id:
            entry = self.name_to_id[name]
            if datetime.now() < entry['expires_at']:
                return entry['id']
            else:
                del self.name_to_id[name]
        return None

    def set_character_id(self, name: str, char_id: int) -> None:
        """存储角色名和ID的映射"""
        self.name_to_id[name] = {
            'id': char_id,
            'expires_at': datetime.now() + self.ttl
        }

    def get_character_affiliation(self, char_id: int) -> Optional[Dict[str, Any]]:
        """获取角色的公司和联盟信息"""
        if char_id in self.id_to_info:
            entry = self.id_to_info[char_id]
            if datetime.now() < entry['expires_at']:
                return entry['info']
            else:
                del self.id_to_info[char_id]
        return None

    def set_character_affiliation(self, char_id: int, info: Dict[str, Any]) -> None:
        """存储角色的公司和联盟信息"""
        self.id_to_info[char_id] = {
            'info': info,
            'expires_at': datetime.now() + self.ttl
        }

    def get_name_for_id(self, entity_id: int) -> Optional[Dict[str, str]]:
        """根据ID获取名称和类别"""
        if entity_id in self.id_to_name:
            entry = self.id_to_name[entity_id]
            if datetime.now() < entry['expires_at']:
                return entry['data']
            else:
                del self.id_to_name[entity_id]
        return None

    def set_name_for_id(self, entity_id: int, data: Dict[str, str]) -> None:
        """存储ID与名称、类别的映射"""
        self.id_to_name[entity_id] = {
            'data': data,
            'expires_at': datetime.now() + self.ttl
        }

    def filter_cached_names(self, names: List[str]) -> List[str]:
        """过滤出未缓存的角色名"""
        now = datetime.now()
        return [name for name in names if name not in self.name_to_id or
                now > self.name_to_id[name]['expires_at']]

    def filter_cached_ids(self, ids: List[int]) -> List[int]:
        """过滤出未缓存的ID"""
        now = datetime.now()
        return [id_ for id_ in ids if id_ not in self.id_to_info or
                now > self.id_to_info[id_]['expires_at']]

    def filter_cached_entity_ids(self, ids: List[int]) -> List[int]:
        """过滤出未缓存的实体ID"""
        now = datetime.now()
        return [id_ for id_ in ids if id_ not in self.id_to_name or
                now > self.id_to_name[id_]['expires_at']]

    def get_cache_stats(self) -> Dict[str, int]:
        """获取缓存统计信息，包括每种缓存的当前条目数"""
        now = datetime.now()

        valid_name_to_id = sum(1 for entry in self.name_to_id.values() if now < entry['expires_at'])
        valid_id_to_info = sum(1 for entry in self.id_to_info.values() if now < entry['expires_at'])
        valid_id_to_name = sum(1 for entry in self.id_to_name.values() if now < entry['expires_at'])

        return {
            'name_to_id': valid_name_to_id,
            'id_to_info': valid_id_to_info,
            'id_to_name': valid_id_to_name,
            'total': valid_name_to_id + valid_id_to_info + valid_id_to_name
        }

    def __repr__(self):
        """打印当前缓存统计信息"""
        stats = self.get_cache_stats()
        msg = ""
        msg += f"角色缓存统计:\n"
        msg += f"  角色名称 -> ID映射: {stats['name_to_id']}条\n"
        msg += f"  ID -> 角色信息映射: {stats['id_to_info']}条\n"
        msg += f"  ID -> 名称/类别映射: {stats['id_to_name']}条\n"
        msg += f"  总缓存条目: {stats['total']}条\n"
        msg += f"  缓存有效期: {self.ttl.days}天"
        return msg


character_cache = CharacterCache()
