# EVE Online DScan Tool

This project is a DScan analysis tool for EVE Online game, used to analyze in-game player affiliations and ship scanning results.

## Features

- **Member Affiliation Analysis**: Analyze the distribution of player affiliations in game
- **Ship Scan Analysis**: Parse and display regular ships, flagships and structure classifications from scans
- **Entity Relationships**: Display relationships between characters, corporations and alliances
- **Interactive Filtering**: Filter relevant entities by clicking or hovering
- **Clean Interface**: Clearly display scan results and related information
- **Short Link Sharing**: Generate shareable short links for easy sharing of analysis results with fleet members

## Direct Usage

Visit [dscan.icu](https://dscan.icu/) to use directly

## Private Deployment

### SDE Data Download

This tool relies on a third-party maintained EVE Online database. For first-time use, please go to [Fuzzwork](https://www.fuzzwork.co.uk/dump/) to download the latest SDE database file, extract it, and place it in the project root directory.

#### Automatic SDE Download

You can automatically download and extract the latest SDE database with the following commands:

```bash
wget https://www.fuzzwork.co.uk/dump/sqlite-latest.sqlite.bz2
bunzip2 sqlite-latest.sqlite.bz2
```

Ensure the extracted file is named `sqlite-latest.sqlite` and placed in the project root directory, or modify the `SQLITE_DB_PATH` setting in `config.py`.

### Standard Deployment

1. Clone the repository:
   ```bash
   git clone https://github.com/zifox/eve-dscan-tool.git
   cd dscan
   ```

2. Install dependencies:
   ```bash
   pip install -r requirements.txt
   ```

3. Download SDE and extract to dscan/:
   ```bash
   wget https://www.fuzzwork.co.uk/dump/sqlite-latest.sqlite.bz2
   bunzip2 sqlite-latest.sqlite.bz2
   ```

4. Start the application:
   ```bash
   uvicorn main:app --host 0.0.0.0 --port 8000
   ```

5. Visit `http://localhost:8000` to begin using

### Docker Deployment

1. Build Docker image:
   ```bash
   docker build -t dscan .
   ```

2. Download SDE data:
   ```bash
   wget https://www.fuzzwork.co.uk/dump/sqlite-latest.sqlite.bz2
   bunzip2 sqlite-latest.sqlite.bz2
   ```

3. Run container:
   ```bash
   docker run -d -p 8000:8000 \
      -v $(pwd)/dscan.sqlite:/app/dscan.sqlite \
      -v $(pwd)/sqlite-latest.sqlite:/app/sqlite-latest.sqlite \
      dscan
   ```

4. Visit `http://localhost:8000` to begin using

### Environment Variables

The following parameters can be configured through environment variables or a `.env` file:

- `DATABASE_URL`: Database connection URL, default is `sqlite+aiosqlite:///./dscan.sqlite`
- `SQLITE_DB_PATH`: EVE SDE database path, default is `./sqlite-latest.sqlite`
- `HTTP_PROXY`/`HTTPS_PROXY`: Proxy server settings (if needed)

## Internationalization Support

This tool supports the following languages:
- Chinese (default)
- English

Language selection is determined by the following priority:
1. User's saved language preference (localStorage)
2. Language settings in browser cookies
3. Browser language
4. Default is Chinese

You can manually change the language through the language switch button in the interface.

### Participate in Internationalization

Internationalization files are located in `static/js/translations.js`, you can add support for new languages by modifying this file.

<details>
  <summary>Translation Example</summary>

```javascript
const translations = {
  'zh': {
    'title': 'EVE Online DScan 分析工具',
    'tool_desc': '这是一个简单的EVE Online DScan分析工具，可以帮助您分析本地频道成员和舰船扫描结果。只需将DScan数据粘贴到下面的文本框中，然后点击提交按钮即可。',
    'paste_dscan_data': '粘贴Dscan数据',
    // More Chinese translations...
  },
  'en': {
    'title': 'EVE Online DScan Analysis Tool',
    'tool_desc': 'This is a simple EVE Online DScan analysis tool that can help you analyze local channel members and ship scan results.Just paste the DScan data into the text box below and click the submit button.',
    'paste_dscan_data': 'Paste DScan Data',
    // More English translations...
  },
  // Add new languages, for example:
  // 'ru': { ... },
  // 'de': { ... }
}
```
</details>

To add a new language, please add corresponding translation key-value pairs following the existing format, then submit a Pull Request.

## TODO List

- [ ] RESTful API interface
- [ ] API key authentication mechanism
- [ ] Support for saving historical scan results
- [ ] Support for more languages

## License

This project is licensed under the [GNU General Public License v3.0 (GPL-3.0)](https://www.gnu.org/licenses/gpl-3.0.html).

## Contribution

Issues and Pull Requests are welcome!