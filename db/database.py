from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
from sqlalchemy.pool import NullPool

from config import settings

# 创建异步数据库引擎
async_engine = create_async_engine(
    settings.DATABASE_URL,
    echo=False,
    future=True,
    connect_args={"check_same_thread": False},
    poolclass=NullPool,
)

# 创建异步会话工厂
AsyncSessionLocal = sessionmaker(
    bind=async_engine,
    expire_on_commit=False,
    class_=AsyncSession,
)

# 创建基础模型类
Base = declarative_base()

# 获取数据库会话
async def get_db():
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()

# 初始化数据库
async def init_db():
    # 创建数据库表
    async with async_engine.begin() as conn:
        # 导入所有模型以确保它们被注册
        from db.models.dscan import LocalDScan, ShipDScan
        from db.models.log import RequestLog
        await conn.run_sync(Base.metadata.create_all)

    print("Database initialized successfully")