# EVE Online DScan 工具

本项目是一个 EVE Online 游戏的 DScan 分析工具，用于分析游戏中的成员势力和舰船分析扫描结果。

## 功能特点

- **成员势力分析**：分析游戏中成员势力的分布情况
- **舰船扫描分析**：解析并展示舰船扫描的常规，旗舰和建筑分类
- **实体关联**：显示角色、军团和联盟之间的关系
- **交互式筛选**：通过点击或悬停可筛选相关实体
- **简明界面**：清晰展示扫描结果和相关信息
- **短链接分享**：生成可分享的短链接，方便与舰队成员共享分析结果

## 直接使用

访问 [dscan.icu](https://dscan.icu/) 直接使用

## 私有化部署

### SDE 数据下载

本工具依赖 EVE Online 的第三方维护数据库。首次使用请前往 [Fuzzwork](https://www.fuzzwork.co.uk/dump/) 下载最新的 SDE 数据库文件并解压，并将其放置在项目根目录下。

#### 自动下载SDE

您可以使用以下命令自动下载并解压最新的SDE数据库：

```bash
wget https://www.fuzzwork.co.uk/dump/sqlite-latest.sqlite.bz2
bunzip2 sqlite-latest.sqlite.bz2
```

确保解压后的文件名为`sqlite-latest.sqlite`并放置在项目根目录，或者修改`config.py`中的`SQLITE_DB_PATH`设置。

### 标准部署

1. 克隆仓库：
   ```bash
   git clone https://github.com/zifox/eve-dscan-tool.git
   cd dscan
   ```

2. 安装依赖：
   ```bash
   pip install -r requirements.txt
   ```
   
3. 下载SDE并解压到dscan/：
   ```bash
   wget https://www.fuzzwork.co.uk/dump/sqlite-latest.sqlite.bz2
   bunzip2 sqlite-latest.sqlite.bz2
   ```

4. 启动应用：
   ```bash
   uvicorn main:app --host 0.0.0.0 --port 8000
   ```

5. 访问 `http://localhost:8000` 开始使用

### Docker 部署

1. 构建 Docker 镜像：
   ```bash
   docker build -t dscan .
   ```
   
2. 下载 SDE 数据：
   ```bash
   wget https://www.fuzzwork.co.uk/dump/sqlite-latest.sqlite.bz2
   bunzip2 sqlite-latest.sqlite.bz2
   ```

3. 运行容器：
   ```bash
   docker run -d -p 8000:8000 \
      -v $(pwd)/dscan.sqlite:/app/dscan.sqlite \
      -v $(pwd)/sqlite-latest.sqlite:/app/sqlite-latest.sqlite \
      dscan
   ```

4. 访问 `http://localhost:8000` 开始使用

### 环境变量

可以通过环境变量或`.env`文件配置以下参数：

- `DATABASE_URL`: 数据库连接URL，默认为`sqlite+aiosqlite:///./dscan.sqlite`
- `SQLITE_DB_PATH`: EVE SDE数据库路径，默认为`./sqlite-latest.sqlite`
- `HTTP_PROXY`/`HTTPS_PROXY`: 代理服务器设置（如需）

## 国际化支持

本工具支持以下语言：
- 中文（默认）
- 英文

语言选择会按以下优先级确定：
1. 用户已保存的语言偏好（localStorage）
2. 浏览器 Cookie 中的语言设置
3. 浏览器语言
4. 默认为中文

可以通过界面的语言切换按钮手动更改语言。

### 参与国际化

国际化文件位于 `static/js/translations.js`，您可以通过修改此文件添加新的语言支持。

<details>
  <summary>翻译示例</summary>
  
```javascript
const translations = {
  'zh': {
    'title': 'EVE Online DScan 分析工具',
    'tool_desc': '这是一个简单的EVE Online DScan分析工具，可以帮助您分析本地频道成员和舰船扫描结果。只需将DScan数据粘贴到下面的文本框中，然后点击提交按钮即可。',
    'paste_dscan_data': '粘贴Dscan数据',
    // 更多中文翻译...
  },
  'en': {
    'title': 'EVE Online DScan Analysis Tool',
    'tool_desc': 'This is a simple EVE Online DScan analysis tool that can help you analyze local channel members and ship scan results.Just paste the DScan data into the text box below and click the submit button.',
    'paste_dscan_data': 'Paste DScan Data',
    // 更多英文翻译...
  },
  // 添加新的语言，例如:
  // 'ru': { ... },
  // 'de': { ... }
}
```
</details>

如需添加新语言，请按照现有格式添加相应的翻译键值对，然后提交 Pull Request。

## TODO 列表

- [ ] RESTful API接口
- [ ] API密钥认证机制
- [ ] 支持保存历史扫描结果
- [ ] 支持更多语言

## 许可证

本项目采用 [GNU通用公共许可证v3.0（GPL-3.0）](https://www.gnu.org/licenses/gpl-3.0.html) 进行许可。

## 贡献

欢迎提交 Issues 和 Pull Requests！