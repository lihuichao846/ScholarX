# ScholarX - 计算机科学前沿论文聚合系统

## 项目简介

ScholarX 是一个专注于计算机科学（CS）领域的论文聚合与分析平台。它旨在帮助研究人员快速获取每日最新的前沿动态、高频研究主题以及关键突破。系统通过聚合多源数据（arXiv, OpenAlex），利用启发式算法生成每日简报，并提供强大的搜索与筛选功能。

## 系统架构

本项目采用前后端分离架构（但在部署上由 Go 服务统一托管静态资源），后端基于 Go (Gin) 构建，前端采用原生 HTML/CSS/JavaScript。

### 目录结构

```
d:\Program Files\文献爬取
├── internal/
│   ├── analysis/       # 核心分析层：每日摘要生成、趋势提取
│   ├── api/            # API 路由处理层
│   ├── model/          # 数据模型定义
│   ├── pkg/
│   │   └── translator/ # 中英学术术语翻译工具
│   └── provider/       # 数据源适配层 (ArXiv, OpenAlex, Semantic Scholar)
├── static/             # 前端静态资源 (HTML, CSS, JS)
├── main.go             # 程序入口与路由注册
└── go.mod              # 依赖管理
```

## 核心模块详解

### 1. 入口与路由 (Main & API)

-   **入口 (`main.go`)**：初始化 Gin 引擎，注册静态文件服务（`/static`）和 API 路由（`/search`, `/daily-summary`）。
-   **处理器 (`internal/api/handlers.go`)**：
    -   `GetDailySummary`：并发拉取 ArXiv 和 OpenAlex 近 48 小时数据，调用分析层生成简报。
    -   `SearchPapers`：处理搜索请求，支持关键词翻译、多源并发检索、S2 引用量回填、CCF 等级筛选及严格日期过滤。

### 2. 数据提供层 (Providers)

位于 `internal/provider/`，负责与外部学术 API 交互并标准化数据。

-   **ArXiv (`arxiv.go`)**：通过 Atom API 获取论文，解析 XML 并处理特殊命名空间字段。
-   **OpenAlex (`openalex.go`)**：利用 Works API 获取论文，解析倒排索引摘要。
-   **Semantic Scholar (`semanticscholar.go`)**：专门用于批量回填论文的引用量数据（解决 arXiv/OpenAlex 引用更新滞后问题）。
-   **通用工具 (`common.go`)**：维护 CCF 会议/期刊分级目录（A/B/C 类），提供日期解析与顶刊顶会过滤逻辑。

### 3. 分析引擎 (Analysis)

位于 `internal/analysis/analyzer.go`，是“每日摘要”功能的核心。

-   **趋势检测**：基于预定义的关键术语（如 LLM, Diffusion, Agent 等）统计词频，识别当日热门领域。
-   **突破识别**：通过关键词匹配（"state-of-the-art", "outperform"）及 CCF A 类标识，筛选高价值论文。
-   **摘要生成**：自动提取论文摘要中的贡献声明（"We propose..."），生成简短的一句话介绍（One-Liner）。
-   **数据结构**：生成包含趋势列表、高频主题词云、精选亮点论文的 `DailySummary` 对象。

### 4. 辅助工具 (Utils)

-   **翻译器 (`internal/pkg/translator/`)**：维护 CS 专业术语的中英映射字典（如 "人工智能" -> "Artificial Intelligence"），支持搜索关键词的自动转换。

### 5. 前端展示 (Frontend)

位于 `static/` 目录，采用轻量级原生实现。

-   **交互逻辑 (`app.js`)**：
    -   **每日摘要**：首页加载时自动拉取并渲染双栏布局的每日简报。
    -   **搜索与筛选**：支持按时间、来源、CCF 等级、顶会进行筛选；支持无限滚动加载。
    -   **动态交互**：实现趋势文本的展开/收起、点击主题词自动搜索等功能。
-   **样式 (`styles.css`)**：基于 Flexbox/Grid 的响应式设计，适配桌面与移动端。

## 关键特性

1.  **每日自动简报**：自动聚合近 48 小时论文，生成“关键趋势”、“精选亮点”和“高频主题”。
2.  **智能中英映射**：支持使用中文搜索 CS 术语，系统自动映射到对应的英文关键词进行检索。
3.  **多源数据融合**：并发请求 arXiv 和 OpenAlex，去重并整合数据；利用 Semantic Scholar 修正引用数据。
4.  **高价值筛选**：内置 CCF A/B/C 类分级标记，支持仅查看顶刊/顶会论文。
5.  **精准引用追踪**：针对 arXiv 论文 ID 进行清洗与版本号处理，确保从 Semantic Scholar 获取准确的引用计数。

## 技术栈

-   **后端**: Go 1.23+, Gin Web Framework
-   **前端**: HTML5, CSS3, Vanilla JavaScript (ES6+)
-   **外部 API**: ArXiv API, OpenAlex API, Semantic Scholar Graph API
