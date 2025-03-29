import httpx
from config import settings

# 全局HTTP客户端
http_client = None

# 初始化HTTP客户端
async def init_http_client():
    global http_client
    
    # 设置代理
    proxies = None
    if settings.HTTP_PROXY or settings.HTTPS_PROXY:
        proxies = {}
        if settings.HTTP_PROXY:
            proxies["http://"] = settings.HTTP_PROXY
        if settings.HTTPS_PROXY:
            proxies["https://"] = settings.HTTPS_PROXY
    
    # 创建客户端
    http_client = httpx.AsyncClient(
        timeout=settings.HTTP_TIMEOUT,
        limits=httpx.Limits(max_connections=settings.HTTP_MAX_CONNECTIONS),
        proxies=proxies,
        headers={
            "Accept": "application/json",
            "Accept-Language": settings.ESI_LANGUAGE,
            "Cache-Control": "no-cache",
        }
    )
    
    print("HTTP client initialized successfully")

# 关闭HTTP客户端
async def close_http_client():
    global http_client
    if http_client:
        await http_client.aclose()
        http_client = None
        print("HTTP client closed successfully")

# 获取HTTP客户端
def get_http_client():
    global http_client
    if not http_client:
        raise RuntimeError("HTTP client not initialized")
    return http_client