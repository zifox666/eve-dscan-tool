FROM python:3.11-slim

WORKDIR /app

# 复制必要文件
COPY requirements.txt .
COPY . .

# 安装依赖
RUN pip install --no-cache-dir -r requirements.txt

# 创建数据目录
RUN mkdir -p /data/sqlite

# 修正暴露端口和启动命令
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
