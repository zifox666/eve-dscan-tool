from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware

import time
from db.database import get_db


class RequestLoggingMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        # 记录请求开始时间
        start_time = time.time()

        # 处理请求
        response = await call_next(request)

        # 计算处理时间
        process_time = time.time() - start_time

        # 异步记录请求日志
        try:
            # 获取数据库会话
            async for db in get_db():
                from db.models.log import RequestLog

                # 创建日志记录
                log_entry = RequestLog(
                    client_ip=request.client.host,
                    request_path=request.url.path,
                    request_method=request.method,
                    process_time=process_time,
                    status_code=response.status_code
                )

                # 添加并提交到数据库
                db.add(log_entry)
                await db.commit()
                break
        except Exception as e:
            # 记录错误但不影响主请求
            print(f"记录请求日志时出错: {str(e)}")

        return response


# 中间件记录真实IP
class CloudflareIPMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request: Request, call_next):
        if "CF-Connecting-IP" in request.headers:
            # 直接覆盖客户端 IP 信息
            request.scope["client"] = (
                request.headers["CF-Connecting-IP"],
                request.scope.get("client")[1] if request.scope.get("client") else 0
            )
        return await call_next(request)