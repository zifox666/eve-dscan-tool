from datetime import datetime, timedelta
from typing import Dict, Any, Optional


class Cache:
    def __init__(self, max_size: int = 128, ttl: int = 3600):
        self.cache: Dict[str, Dict[str, Any]] = {}
        self.max_size = max_size
        self.ttl = ttl

    def get(self, key: str) -> Optional[Dict[str, Any]]:
        if key not in self.cache:
            return None

        item = self.cache[key]
        if datetime.now() > item['expires_at']:
            del self.cache[key]
            return None

        return item['data']

    def set(self, key: str, value: Dict[str, Any]) -> None:
        # 清理过期项
        self._cleanup()

        # 如果缓存已满，删除最早的项
        if len(self.cache) >= self.max_size:
            oldest_key = min(self.cache.items(), key=lambda x: x[1]['expires_at'])[0]
            del self.cache[oldest_key]

        self.cache[key] = {
            'data': value,
            'expires_at': datetime.now() + timedelta(seconds=self.ttl)
        }

    def _cleanup(self) -> None:
        now = datetime.now()
        expired_keys = [k for k, v in self.cache.items() if now > v['expires_at']]
        for k in expired_keys:
            del self.cache[k]


# 创建全局缓存实例
dscan_cache = Cache()