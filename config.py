import os

from dotenv import load_dotenv


load_dotenv()


class Settings:
    # 应用设置
    APP_NAME = "EVE Online DScan Tool"
    APP_VERSION = "1.1.0"
    APP_DESCRIPTION = "A tool for analyzing EVE Online DScan data"

    # 储存数据 数据库设置
    DATABASE_URL = os.getenv("DATABASE_URL", "sqlite+aiosqlite:///./dscan.sqlite")

    # EVE ESI API设置
    ESI_BASE_URL = "https://esi.evetech.net/latest"
    ESI_DATASOURCE = "tranquility"
    ESI_LANGUAGE = "en"

    # HTTP客户端设置
    HTTP_TIMEOUT = 30.0
    HTTP_MAX_CONNECTIONS = 100
    HTTP_PROXY = os.getenv("HTTP_PROXY", None)
    HTTPS_PROXY = os.getenv("HTTPS_PROXY", None)

    # 短链接设置
    SHORT_LINK_LENGTH = 10

    # 静态资源
    STATIC_DIR = "static"
    TEMPLATES_DIR = "templates"

    # 每批处理的最大角色数量（EVE API限制）
    MAX_CHARACTERS_PER_BATCH = 1000
    # 角色缓存天数
    TTL_CHARACTER_CACHE_DAYS = 7

    # EVE SDE 只读数据库设置
    SQLITE_DB_PATH = os.getenv("SQLITE_DB_EN_PATH", "./sqlite-latest.sqlite")


settings = Settings()
