from sqlalchemy import Column, Integer, String, Text, DateTime, ForeignKey, Boolean, JSON
from sqlalchemy.orm import relationship
from datetime import datetime

from db.database import Base

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