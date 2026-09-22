# UICL

**Universal Intent and Contract Language · 通用意图与契约语言**

UICL 是用于描述意图、内容、资源、应用与执行契约的开放语言设计，标准文件后缀为 `.uicl`。Semaquil 是一种实现品牌，不是该格式的唯一入口。

本仓库收录 UICL 1.0 规范整合稿、20 个领域 Profile、42 份示例文档，以及 Python 参考解析与检查工具。版本 1.0 指规范整合版本，不代表完整编译器、运行时或云部署适配器已经实现。

## 阅读入口

- [在线文档](https://payhon.github.io/uicl/)
- [完整使用说明](README.zh-CN.md)
- [语言规范](spec/UICL-1.0.md) · [UICL 原生文档](spec/UICL-1.0.uicl)
- [结构化单文件全集](UICL-1.0-Structured.uicl)
- [快速入门](spec/Quickstart.md)
- [Profile 索引](spec/Profile-Index.md) · [Profile 作者指南](spec/Profile-Authoring.md)
- [完整应用示例](examples/fullstack/project.uicl) · [应用示例说明](spec/Fullstack-Walkthrough.md)
- [编译器与一致性范围](spec/Compiler-Conformance.md)

## 快速示例

```uicl
uicl "1.0"

app "你好"
  ui.text "我的第一个 UICL 应用"
```

## 本地验证

需要 Python 3.10 或更高版本。Markdown 宿主检查另需安装已固定的依赖。

```bash
python -m pip install -r requirements-host.txt
python tools/check.py --out reports/structural-check.json
python tools/check.py examples/fullstack/project.uicl
python tools/host.py --out reports/host-check.json
python tools/rebuild_views.py --check
python -m unittest discover -s tests -v
python tools/verify_package.py
```

工具只执行其说明的语法、结构、引用与有限静态检查。不会连接示例服务器、调用模型、修改数据库或部署应用。完整类型与效果验证、UI/后端/AIGC/Agent/Wasm 运行时、平台构建和部署适配器尚未实现。

## 源码与派生文件

### 文档站开发

文档站使用 Rspress，需要 Node.js 24+ 和 pnpm 12.5.1。

```bash
pnpm install
pnpm dev
pnpm check
pnpm preview
```

开发及预览地址以命令输出为准，站点路径为 `/uicl/`。`pnpm check` 检查内容接入并构建静态站点，输出到 `doc_build/`。

首页与工具页维护在 `website/pages/`，主题在 `theme/`，静态资源在 `website/public/`。`scripts/prepare-docs.mjs` 从现有规范、Profile 指南与示例生成 `website/docs/`，请勿直接修改生成目录。修改规范源后重新运行 `pnpm dev`，或运行 `pnpm build` 生成发布版本。

GitHub Pages 使用 `.github/workflows/docs.yml` 发布：PR 执行检查，推送 `main` 后构建并部署。首次发布需在仓库 **Settings → Pages → Source** 选择 **GitHub Actions**。站点固定使用 `/uicl/` 前缀；更换仓库名或域名时同步修改 `rspress.config.ts` 的 `base` 和 `siteOrigin`。

### 语言规范派生文件

`spec/`、`profiles/`、`meta/grammar.uicl` 和 `examples/` 是各自的维护入口。单文件全集、文法阅读视图和示例索引由工具派生，不应另行手工编辑。修改规范后运行 `python tools/rebuild_views.py` 并重新验证。

## 开放与授权边界

UICL 格式与 Semaquil 产品品牌分离。授权方案及尚待正式落实的事项以 [完整使用说明](README.zh-CN.md#维护与开放边界) 为准；此次导入未更改授权条款，也未重新许可 `provenance/` 中的第三方资料。用户文件的内容权利不因使用 `.uicl` 后缀而改变。
