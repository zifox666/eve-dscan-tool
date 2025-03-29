# db/sde.py
import sqlite3
from typing import List, Dict, Any, Optional
from functools import lru_cache
from config import settings


class EVESqliteDB:
    """EVE SDE SQLite数据库访问类，支持多语言"""

    def __init__(self):
        """初始化多语言EVE SQLite数据库连接"""
        self.db_connections = {
            "zh": None,
            "en": None
        }
        self.cursors = {
            "zh": None,
            "en": None
        }
        # 初始化连接
        self.connect()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def connect(self):
        """连接到所有语言的EVE SDE SQLite数据库"""
        try:
            # 中文数据库连接
            self.db_connections["zh"] = sqlite3.connect(settings.SQLITE_DB_ZH_PATH)
            self.db_connections["zh"].row_factory = sqlite3.Row
            self.cursors["zh"] = self.db_connections["zh"].cursor()

            # 英文数据库连接
            self.db_connections["en"] = sqlite3.connect(settings.SQLITE_DB_EN_PATH)
            self.db_connections["en"].row_factory = sqlite3.Row
            self.cursors["en"] = self.db_connections["en"].cursor()

            return True
        except Exception as e:
            print(f"连接EVE SDE数据库失败: {e}")
            return False

    def close(self):
        """关闭所有数据库连接"""
        for lang in self.cursors:
            if self.cursors[lang]:
                self.cursors[lang].close()
        for lang in self.db_connections:
            if self.db_connections[lang]:
                self.db_connections[lang].close()

    @lru_cache(maxsize=1024)
    def get_type_info(self, type_id: int, language: str = "zh") -> Optional[Dict[str, Any]]:
        """根据type_id获取指定语言的类型信息

        Args:
            type_id: 物品类型ID
            language: 语言代码，支持"zh"和"en"，默认为"zh"

        Returns:
            包含类型信息的字典，如果未找到则返回None
        """
        if language not in self.cursors:
            language = "zh"  # 默认回退到中文

        cursor = self.cursors[language]
        if not cursor:
            return None

        try:
            # 从新的types表获取信息
            cursor.execute("""
                SELECT type_id, name, group_name, category_name, groupID, categoryID
                FROM types
                WHERE type_id = ?
            """, (type_id,))

            row = cursor.fetchone()
            if not row:
                return None

            result = dict(row)
            return result

        except Exception as e:
            print(f"查询类型信息失败 ({language}): {e}")
            return None

    def get_type_infos(self, type_ids: List[int], language: str = "zh") -> Dict[int, Dict[str, Any]]:
        """批量获取多个type_id的类型信息

        Args:
            type_ids: 类型ID列表
            language: 语言代码，支持"zh"和"en"，默认为"zh"

        Returns:
            以type_id为键，类型信息为值的字典
        """
        result = {}
        if not type_ids:
            return result

        # 逐个获取信息，利用缓存提高性能
        for type_id in type_ids:
            info = self.get_type_info(type_id, language)
            if info:
                result[type_id] = info

        return result


# 创建全局数据库实例
eve_db = EVESqliteDB()