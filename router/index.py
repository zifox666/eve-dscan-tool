from fastapi import APIRouter, Request, Form
from fastapi.responses import HTMLResponse
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
async def submit_dscan(
        request: Request,
        dscan_data: str = Form(...),
        filter_distance: bool = Form(False),
        language: str = Form(None)
):
    """提交DScan数据，根据类型重定向到相应的处理页面"""
    # 检测DScan类型
    dscan_type = detect_dscan_type(dscan_data)

    # 创建表单数据
    form_data = {"data": dscan_data}
    if filter_distance:
        form_data["filter_distance"] = "true"

    if language:
        form_data["language"] = language

    # 根据类型重定向到POST处理路由
    if dscan_type == "local":
        return templates.TemplateResponse(
            "redirect_form.html.jinja2",
            {
                "request": request,
                "action": "/c/process",
                "form_data": form_data
            }
        )
    else:  # ship类型
        return templates.TemplateResponse(
            "redirect_form.html.jinja2",
            {
                "request": request,
                "action": "/v/process",
                "form_data": form_data
            }
        )