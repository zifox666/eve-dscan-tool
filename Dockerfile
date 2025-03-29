FROM python:3.11-slim

WORKDIR /app

# 复制必要文件
COPY requirements.txt .
# 安装依赖
RUN pip install --no-cache-dir -r requirements.txt
COPY . .

VOLUME ["/app/dscan.sqlite", "/app/item_db_zh.sqlite"]
# 修正暴露端口和启动命令
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
