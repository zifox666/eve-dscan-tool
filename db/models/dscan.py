import json
from datetime import datetime

from sqlalchemy import Column, Integer, String, Text, DateTime, JSON
from sqlalchemy.types import TypeDecorator

from db.database import Base


class JSONEncodedDict(TypeDecorator):
    """将JSON结构表示为SQLite中的TEXT列"""
    impl = Text

    def process_bind_param(self, value, dialect):
        if value is not None:
            value = json.dumps(value)
        return value

    def process_result_value(self, value, dialect):
        if value is not None:
            value = json.loads(value)
        return value


class LocalDScan(Base):
    __tablename__ = "local_dscan"

    id = Column(Integer, primary_key=True, index=True)
    short_id = Column(String(10), unique=True, index=True, nullable=False)
    raw_data = Column(Text, nullable=False)  # 原始上传数据
    processed_data = Column(JSON, nullable=False)  # 处理后的数据
    client_ip = Column(String(50), nullable=False)  # 创建者IP
    view_count = Column(Integer, default=0)  # 访问次数
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    def __repr__(self):
        return f"<LocalDScan(id={self.id}, short_id={self.short_id})>"


class ShipDScan(Base):
    __tablename__ = "ship_dscan"

    id = Column(Integer, primary_key=True, index=True)
    short_id = Column(String(10), unique=True, index=True, nullable=False)
    raw_data = Column(Text, nullable=False)  # 原始上传数据
    processed_data = Column(JSON, nullable=False)  # 处理后的数据
    client_ip = Column(String(50), nullable=False)  # 创建者IP
    view_count = Column(Integer, default=0)  # 访问次数
    created_at = Column(DateTime, default=datetime.utcnow)
    updated_at = Column(DateTime, default=datetime.utcnow, onupdate=datetime.utcnow)

    def __repr__(self):
        return f"<ShipDScan(id={self.id}, short_id={self.short_id})>"
