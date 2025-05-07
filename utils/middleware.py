from fastapi import Request
from starlette.middleware.base import BaseHTTPMiddleware

import time
from db.database import get_db


class RequestLoggingMiddleware(BaseHTTPMiddleware):
    """中间件 记录请求日志"""
    async def dispatch(self, request: Request, call_next):
        start_time = time.time()

        response = await call_next(request)

        process_time = time.time() - start_time

        try:
            async for db in get_db():
                from db.models.log import RequestLog

                log_entry = RequestLog(
                    client_ip=request.client.host,
                    request_path=request.url.path,
                    request_method=request.method,
                    process_time=process_time,
                    status_code=response.status_code
                )

                db.add(log_entry)
                await db.commit()
                break
        except Exception as e:
            print(f"记录请求日志时出错: {str(e)}")

        return response


class CloudflareIPMiddleware(BaseHTTPMiddleware):
    """中间件 获取Cloudflare CDN的客户端真实IP"""
    async def dispatch(self, request: Request, call_next):
        if "CF-Connecting-IP" in request.headers:
            # 直接覆盖客户端 IP 信息
            request.scope["client"] = (
                request.headers["CF-Connecting-IP"],
                request.scope.get("client")[1] if request.scope.get("client") else 0
            )
        return await call_next(request)