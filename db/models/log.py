from datetime import datetime

from sqlalchemy import Column, Integer, String, Float, DateTime

from db.database import Base


class RequestLog(Base):
    __tablename__ = "request_logs"

    id = Column(Integer, primary_key=True, index=True)
    client_ip = Column(String(50), index=True, nullable=False)  # 客户端IP
    request_path = Column(String(255), nullable=False)  # 请求路径
    request_method = Column(String(10), nullable=False)  # 请求方法
    process_time = Column(Float, nullable=False)  # 处理时间
    status_code = Column(Integer, nullable=False)  # 状态码
    created_at = Column(DateTime, default=datetime.utcnow)  # 使用Python的datetime

    def __repr__(self):
        return f"<RequestLog(id={self.id}, client_ip={self.client_ip}, path={self.request_path})>"
