from contextlib import asynccontextmanager

import uvicorn
from fastapi import FastAPI
from fastapi.staticfiles import StaticFiles

# 导入自定义模块
from api.client import init_http_client
from db.database import init_db
from router.index import router as index_router
from router.local_dscan import router as local_dscan_router
from router.ship_dscan import router as ship_dscan_router
from utils.middleware import CloudflareIPMiddleware, RequestLoggingMiddleware


@asynccontextmanager
async def lifespan(app: FastAPI):
    # 初始化数据库连接
    await init_db()
    # 初始化HTTP客户端
    await init_http_client()
    yield
    # 关闭HTTP客户端
    from api.client import close_http_client
    await close_http_client()

# 创建FastAPI应用
app = FastAPI(title="EVE Online DScan Tool", lifespan=lifespan)

# 添加中间件
app.add_middleware(CloudflareIPMiddleware)
app.add_middleware(RequestLoggingMiddleware)

# 挂载静态文件
app.mount("/static", StaticFiles(directory="static"), name="static")

# 注册路由
app.include_router(index_router)
app.include_router(local_dscan_router, prefix="/c")
app.include_router(ship_dscan_router, prefix="/v")

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=8000, reload=True)