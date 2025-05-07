import sqlite3
from typing import List, Dict, Any, Optional

class SQLiteHelper:
    """SQLite数据库助手类，用于连接和查询EVE数据库"""
    
    def __init__(self, db_path: str):
        """初始化SQLite连接
        
        Args:
            db_path: SQLite数据库文件路径
        """
        self.db_path = db_path
        self.conn = None
        self.cursor = None
    
    def connect(self):
        """连接到SQLite数据库"""
        try:
            self.conn = sqlite3.connect(self.db_path)
            self.conn.row_factory = sqlite3.Row  # 使结果可以通过列名访问
            self.cursor = self.conn.cursor()
            return True
        except Exception as e:
            print(f"连接数据库失败: {e}")
            return False
    
    def close(self):
        """关闭数据库连接"""
        if self.cursor:
            self.cursor.close()
        if self.conn:
            self.conn.close()
    
    def get_type_info(self, type_id: int) -> Optional[Dict[str, Any]]:
        """根据type_id获取类型信息
        
        Args:
            type_id: 类型ID
            
        Returns:
            包含类型信息的字典，如果未找到则返回None
        """
        try:
            self.cursor.execute(
                """SELECT type_id, name, zh_name, group_name, category_name 
                   FROM types WHERE type_id = ?""", 
                (type_id,)
            )
            row = self.cursor.fetchone()
            if row:
                return dict(row)
            return None
        except Exception as e:
            print(f"查询类型信息失败: {e}")
            return None
    
    def get_type_infos(self, type_ids: List[int]) -> Dict[int, Dict[str, Any]]:
        """批量获取多个type_id的类型信息
        
        Args:
            type_ids: 类型ID列表
            
        Returns:
            以type_id为键，类型信息为值的字典
        """
        result = {}
        if not type_ids:
            return result
            
        try:
            # 构建IN查询的参数占位符
            placeholders = ",".join(["?" for _ in type_ids])
            query = f"""SELECT type_id, name, zh_name, group_name, category_name 
                      FROM types WHERE type_id IN ({placeholders})"""
            
            self.cursor.execute(query, type_ids)
            rows = self.cursor.fetchall()
            
            for row in rows:
                row_dict = dict(row)
                result[row_dict["type_id"]] = row_dict
                
            return result
        except Exception as e:
            print(f"批量查询类型信息失败: {e}")
            return result