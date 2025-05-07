import json
from typing import Optional

from fastapi import APIRouter, Request, Form, Depends
from fastapi.responses import HTMLResponse, RedirectResponse
from fastapi.templating import Jinja2Templates
from sqlalchemy import select
from sqlalchemy.ext.asyncio import AsyncSession

from api.eve_api import get_character_ids, get_character_affiliations, get_names_for_ids
from config import settings
from db.database import get_db
from db.models.dscan import LocalDScan
from utils.cache import dscan_cache
from utils.helpers import generate_short_id, parse_local_dscan, organize_local_dscan_data, format_time_ago

# 创建路由
router = APIRouter()

# 设置模板
templates = Jinja2Templates(directory=settings.TEMPLATES_DIR)


@router.post("/process", response_class=HTMLResponse)
async def process_local_dscan(
        request: Request,
        data: str = Form(...),
        db: AsyncSession = Depends(get_db),
        language: Optional[str] = Form(None)
):
    """处理Local DScan数据"""
    character_names = parse_local_dscan(data)

    character_ids_map = await get_character_ids(character_names)
    character_ids = list(character_ids_map.values())

    affiliations = await get_character_affiliations(character_ids)

    ids_to_query = set()
    for affiliation in affiliations:
        ids_to_query.add(affiliation["character_id"])
        ids_to_query.add(affiliation["corporation_id"])
        if "alliance_id" in affiliation:
            ids_to_query.add(affiliation["alliance_id"])

    names = await get_names_for_ids(list(ids_to_query))

    result = organize_local_dscan_data(character_names, affiliations, names)

    short_id = generate_short_id(settings.SHORT_LINK_LENGTH)

    client_ip = request.client.host

    new_dscan = LocalDScan(
        short_id=short_id,
        raw_data=data,
        processed_data=result,
        client_ip=client_ip
    )
    db.add(new_dscan)
    await db.commit()
    await db.refresh(new_dscan)

    return RedirectResponse(url=f"/c/{short_id}", status_code=303)


@router.get("/{short_id}", response_class=HTMLResponse)
async def view_local_dscan(
        request: Request,
        short_id: str,
        db: AsyncSession = Depends(get_db),
):
    """查看保存的Local DScan数据"""
    cache_key = f"local_dscan:{short_id}"
    cached_data = dscan_cache.get(cache_key)

    query = select(LocalDScan).where(LocalDScan.short_id == short_id)
    result = await db.execute(query)
    dscan = result.scalar_one_or_none()

    if not dscan:
        return templates.TemplateResponse(
            "404.html",
            {"request": request, "message": "找不到指定的本地扫描数据"}
        )

    dscan.view_count += 1
    await db.commit()

    if cached_data:
        if isinstance(cached_data, str):
            dscan_data = json.loads(cached_data)
        else:
            dscan_data = cached_data
        dscan_data["view_count"] = dscan.view_count
    else:
        dscan_data = {
            "id": dscan.id,
            "short_id": dscan.short_id,
            "processed_data": dscan.processed_data,
            "view_count": dscan.view_count,
            "created_at": dscan.created_at.isoformat(),
            "time_ago": format_time_ago(dscan.created_at)
        }

    dscan_cache.set(cache_key, dscan_data)
    print(dscan_data)

    return templates.TemplateResponse(
        "local_dscan.html.jinja2",
        {
            "request": request,
            "title": f"Local DScan - Dscan.icu",
            "dscan": dscan_data,
            "base_url": f"{request.base_url}v/{short_id}"
        }
    )
