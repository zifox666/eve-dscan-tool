# db/sde.py
import sqlite3
from functools import lru_cache
from typing import List, Dict, Any, Optional

from config import settings


class EVESqliteDB:
    """EVE SDE SQLite数据库访问类，支持多语言"""

    def __init__(self):
        """初始化EVE SQLite数据库连接"""
        self.db_connection = None
        self.cursor = None
        # 翻译类型常量
        self.TC_TYPE = 8  # 物品类型名称
        self.TC_GROUP = 7  # 分组名称
        self.TC_CATEGORY = 6  # 分类名称

        self.LANGUAGES = {
            "zh": "zh",
            "en": "en"
        }

        self.connect()

    def __enter__(self):
        return self

    def __exit__(self, exc_type, exc_val, exc_tb):
        self.close()

    def connect(self):
        """连接到EVE SDE SQLite数据库"""
        try:
            self.db_connection = sqlite3.connect(settings.SQLITE_DB_PATH)
            self.db_connection.row_factory = sqlite3.Row
            self.cursor = self.db_connection.cursor()
            return True
        except Exception as e:
            print(f"连接EVE SDE数据库失败: {e}")
            return False

    def close(self):
        """关闭数据库连接"""
        if self.cursor:
            self.cursor.close()
        if self.db_connection:
            self.db_connection.close()

    def _get_translation(self, key_id: int, tc_id: int, language: str) -> Optional[str]:
        """获取指定ID的翻译文本

        :param key_id: 键ID（物品ID、分组ID或分类ID）
        :param tc_id: 翻译类别ID (5=type, 7=group, 6=category)
        :param language: 语言代码

        :return: 翻译文本，如果未找到则返回None
        """
        if not self.cursor:
            return None

        try:
            self.cursor.execute("""
                SELECT text FROM trnTranslations 
                WHERE keyID = ? AND tcID = ? AND languageID = ?
            """, (key_id, tc_id, language))

            row = self.cursor.fetchone()
            if row:
                return row["text"]
            return None
        except Exception as e:
            print(f"获取翻译失败: {e}")
            return None

    @lru_cache(maxsize=1024)
    def get_type_info(self, type_id: int, language: str = "zh") -> Optional[Dict[str, Any]]:
        """根据type_id获取指定语言的类型信息

        :param type_id: 物品类型ID
        :param language: 语言代码，支持"zh"和"en"，默认为"zh"

        :return: 包含类型信息的字典，如果未找到则返回None
        """
        if language not in self.LANGUAGES:
            language = "zh"  # 默认回退到中文

        if not self.cursor:
            return None

        try:
            self.cursor.execute("""
                SELECT t.typeID, t.groupID, t.typeName as default_name, 
                       g.groupName as default_group_name, g.categoryID,
                       c.groupName as default_category_name
                FROM invTypes t
                LEFT JOIN invGroups g ON t.groupID = g.groupID
                LEFT JOIN invGroups c ON g.categoryID = c.groupID
                WHERE t.typeID = ?
            """, (type_id,))

            row = self.cursor.fetchone()
            if not row:
                return None

            result = dict(row)

            type_name = self._get_translation(type_id, self.TC_TYPE, language) or result["default_name"]
            group_id = result["groupID"]
            group_name = self._get_translation(group_id, self.TC_GROUP, language) or result["default_group_name"]

            category_id = result["categoryID"]
            category_name = self._get_translation(category_id, self.TC_CATEGORY, language) or result[
                "default_category_name"]

            result["name"] = type_name
            result["group_name"] = group_name
            result["category_name"] = category_name

            return result

        except Exception as e:
            print(f"查询类型信息失败: {e}")
            return None

    def get_type_infos(self, type_ids: List[int], language: str = "zh") -> Dict[int, Dict[str, Any]]:
        """批量获取多个type_id的类型信息

        :param type_ids: 类型ID列表
        :param language: 语言代码，支持"zh"和"en"，默认为"zh"

        :return: 以type_id为键，类型信息为值的字典
        """
        result = {}
        if not type_ids:
            return result

        for type_id in type_ids:
            info = self.get_type_info(type_id, language)
            if info:
                result[type_id] = info

        return result
    

    @lru_cache(maxsize=2048)
    def get_system_id_by_en_name(self, system_name: str) -> Optional[Dict[str, Any]]:
        try:
            self.cursor.execute("""
                SELECT itemID
                FROM mapDenormalize
                WHERE itemName = ?
            """, (system_name,))

            row = self.cursor.fetchone()
            if row:
                return row['itemID']
            return None
        except Exception as e:
            print(f"获取system_id_by_name失败: {e}")
            return None

    @lru_cache(maxsize=2048)
    def get_system_info_by_id(self, system_id: int) -> Optional[Dict[str, Any]]:
        try:
            self.cursor.execute("""
                SELECT itemID, constellationID, regionID, itemName, security
                FROM mapDenormalize
                WHERE itemID = ?
            """, (system_id,))

            row = self.cursor.fetchone()
            if row:
                result = dict(row)

                self.cursor.execute("""
                    SELECT constellationName
                    FROM mapConstellations
                    WHERE constellationID = ?
                """, (result["constellationID"],))
                constellation_row = self.cursor.fetchone()
                if constellation_row:
                    result["constellationName"] = constellation_row["constellationName"]
                else:
                    result["constellationName"] = None

                self.cursor.execute("""
                    SELECT regionName
                    FROM mapRegions
                    WHERE regionID = ?
                """, (result["regionID"],))
                region_row = self.cursor.fetchone()
                if region_row:
                    result["regionName"] = region_row["regionName"]
                else:
                    result["regionName"] = None

                return result
            return None
        except Exception as e:
            print(f"获取翻译失败: {e}")
            return None


eve_db = EVESqliteDB()
