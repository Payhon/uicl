# 工具与验证

UICL 仓库包含规范和有限检查原型。以下命令用于读取和验证契约；当前没有 `uicl dev`、`uicl build` 或 `uicl deploy` 语言运行时命令。

## 准备环境

语言工具使用 Python 3.10+。结构解析、Profile 检查和 HEXL 工具只依赖 Python 标准库；Markdown 宿主检查需要安装仓库中锁定的依赖。

```bash
git clone git@github.com:Payhon/uicl.git
cd uicl
python3 -m venv .venv
source .venv/bin/activate
python -m pip install -r requirements-host.txt
```

## 解析与静态检查

```bash
# 解析计数器源码，不启动应用。
python tools/syntax.py examples/02-counter.uicl --out generated/counter.syntax.json

# 检查 Profile、示例、模块引用和有限静态约束。
python tools/check.py

# 检查完整应用入口及其显式导入。
python tools/check.py examples/fullstack/project.uicl

# 检查 Markdown / HTML 中实际承载的 UICL 文档。
python tools/host.py
```

第三方 Profile 需要显式传入，工具不会自动从网络安装：

```bash
python tools/check.py examples/profile-authoring/measurement-example.uicl \
  --extra-profile examples/profile-authoring/measurement-profile.uicl
```

## 运行仓库检查

```bash
# 确认派生视图与规范源同步，不修改源文件。
python tools/rebuild_views.py --check

# 运行语言工具测试。
python -m unittest discover -s tests -v

# 执行整包检查，并写入 reports/verification.json。
python tools/verify_package.py
```

通过这些检查，不代表 UI、数据库、模型、设备或部署适配器已经运行。有关检查层级和未实现范围，见[编译器与一致性](spec/Compiler-Conformance.md)；真实应用还需要[完整应用示例说明](spec/Fullstack-Walkthrough.md)中列出的宿主绑定。

## 维护文档站

文档站使用 Rspress，需要 Node.js 24+ 和仓库 `package.json` 指定的 pnpm 版本。

```bash
pnpm install
pnpm dev

# 检查内容生成与链接，并构建可发布的静态站点。
pnpm check
```

规范和 Profile 的来源保持在仓库原位置，文档站构建前会检查 Markdown 阅读副本与 `.uicl` 源是否一致，再生成网页。首页和本站说明在 `website/pages/` 维护，`website/docs/` 与 `website/sidebar.json` 是自动生成文件。

修改规范源后，应先同步它的阅读副本和受影响的派生视图，再重新运行文档站命令。不要直接编辑生成目录。
