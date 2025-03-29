from sqlalchemy import create_engine
from sqlalchemy.ext.declarative import declarative_base
from sqlalchemy.orm import sessionmaker
from sqlalchemy.ext.asyncio import create_async_engine, AsyncSession
import redis.asyncio as redis

from config import settings

# 创建异步数据库引擎
async_engine = create_async_engine(
    settings.DATABASE_URL,
    echo=False,
    future=True,
)

# 创建会话工厂
AsyncSessionLocal = sessionmaker(
    bind=async_engine,
    class_=AsyncSession,
    expire_on_commit=False,
)

# 创建基础模型类
Base = declarative_base()

# Redis连接池
redis_pool = None

# 获取数据库会话
async def get_db():
    async with AsyncSessionLocal() as session:
        try:
            yield session
        finally:
            await session.close()

# 获取Redis连接
async def get_redis():
    redis_conn = redis.Redis.from_pool(redis_pool)
    try:
        yield redis_conn
    finally:
        await redis_conn.close()

# 初始化数据库
async def init_db():
    global redis_pool
    
    # 创建数据库表
    async with async_engine.begin() as conn:
        # 导入所有模型以确保它们被注册
        from db.models.dscan import LocalDScan, ShipDScan
        
        # 创建表
        await conn.run_sync(Base.metadata.create_all)
    
    # 初始化Redis连接池
    redis_pool = redis.ConnectionPool.from_url(
        settings.REDIS_URL,
        decode_responses=True
    )
    
    print("Database initialized successfully")