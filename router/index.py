from fastapi import APIRouter, Request, Form, Depends
from fastapi.responses import HTMLResponse, RedirectResponse
from fastapi.templating import Jinja2Templates

from config import settings
from utils.helpers import detect_dscan_type

# 创建路由
router = APIRouter()

# 设置模板
templates = Jinja2Templates(directory=settings.TEMPLATES_DIR)

@router.get("/", response_class=HTMLResponse)
async def index(request: Request):
    """首页"""
    return templates.TemplateResponse(
        "index.html",
        {"request": request, "title": settings.APP_NAME}
    )


@router.post("/submit")
async def submit_dscan(request: Request, dscan_data: str = Form(...), filter_distance: bool = Form(False)):
    """提交DScan数据，根据类型重定向到相应的处理页面"""
    # 检测DScan类型
    dscan_type = detect_dscan_type(dscan_data)

    # 根据类型重定向
    if dscan_type == "local":
        return RedirectResponse(url=f"/c/process?data={dscan_data}", status_code=303)
    else:  # ship类型
        filter_param = "&filter_distance=true" if filter_distance else ""
        return RedirectResponse(url=f"/v/process?data={dscan_data}{filter_param}", status_code=303)