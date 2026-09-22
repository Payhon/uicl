# 工具与验证

UICL 提供 Go 语言核心与 CLI，以及保留的 Python 参考工具。CLI 读取、检查和格式化契约；当前没有 `uicl dev`、`uicl build` 或 `uicl deploy` 语言运行时命令。

## 安装 Go CLI

构建需要 Go 1.24+；构建后的单个二进制内嵌全部标准 Profile，运行不需要 Python、Go 或标准规范目录。

```bash
go build -o bin/uicl ./cmd/uicl
# 或安装到 Go 的可执行文件目录：
go install ./cmd/uicl
```

以下命令假定已将 `uicl` 放入 PATH；也可以替换为 `./bin/uicl`。

## 检查、格式化与查询

```bash
uicl check examples/fullstack/project.uicl --json
uicl check examples/hosted/article.uicl examples/hosted/page.html

# 默认只输出格式化结果；以下两种模式互斥。
uicl fmt examples/02-counter.uicl
uicl fmt --check examples/02-counter.uicl
uicl fmt --write examples/02-counter.uicl

uicl profile inspect uicl.ui --json
uicl explain PRIMARY_CONFLICT
uicl lock --check
uicl version
```

`check` 和 `fmt` 接收明确文件，不递归扫描目录。`.uicl` 后缀也可能承载 Markdown 或 HTML，模式依据真实标记判断；错误版本或冲突标记不会降级为普通 Markdown。

格式化仅调整结构空白，不重排节点/属性、不注入默认值、不转换集合或数字写法。注释、空行、原文块、BOM、换行形式和宿主正文保持原样；编码契约只有在格式变化时重新编码。解析出错不写文件。多文件写入前统一预检，每个文件写入前核对摘要；后续文件失败时，输出中的 `written` 字段说明哪些文件已写入，不承诺跨文件事务，也不能锁住其他程序的任意并发写入。

## 工作区、Profile 与标准输入

`--root` 默认当前目录。入口文件相对当前目录；模块 URI 相对声明文件。导入和显式 Profile 文件在解析符号链接后必须仍位于根目录中。除导入外的资源 URI 只登记，不读取远程资源或解析凭据。

```bash
# 扩展是显式、可重复的参数，不覆盖同名标准定义。
uicl check --extra-profile examples/profile-authoring/measurement-profile.uicl \
  examples/profile-authoring/measurement-example.uicl

# 规范开发时明确检查磁盘上的目录，替换内嵌标准集合。
uicl check --profile-dir profiles examples/fullstack/project.uicl

# --filename 给未保存输入提供身份和相对导入的基址。
cat examples/fullstack/project.uicl | uicl check - \
  --filename examples/fullstack/project.uicl --json
```

标准输入仅接受一个 `-`，必须提供 `--filename`，不能与 `fmt --write` 一起使用。`lock --check` 独立验证现有包源锁中的实际源文件、摘要、版本和依赖；不重写锁、不替代缺失文件、不比较历史 Python 环境。普通 `check` 不要求锁文件存在。

## 输出与能力范围

- 退出码 `0`：声明范围通过，或 `fmt --check` 无变化；`1`：契约/锁检查失败或需格式化；`2`：用法、读取或工具错误。
- `--json` 的 stdout 是一个 `schemaVersion: 1` 对象；诊断按文件、字节位置和错误码排序。
- JSON 范围是零基、右端不包含的字节偏移；行列零基，列按 Unicode 标量计数。终端显示一基行列，后续 LSP 需要转换位置编码。
- 编码契约的诊断指向载体范围，并附带解码后的位置；无法唯一定位节点的跨节点诊断使用解码后整个载荷的范围。普通契约保留逐字节源码映射。
- 已实现语法、Profile 形状、模块与身份/端口链接，以及列明的有限领域检查。完整运行绑定类型、泛型端口、权限与效果、recipe 展开、真实执行和产物验收仍为 `NOT_EVALUATED`。
- 宿主支持 CommonMark 注解与 HTML 契约提取/局部修改；不执行 JavaScript，不宣称浏览器渲染或全部宿主一致性已经通过。

根 Go 包可供后续工具复用：`Parse`/`ParseCore` 返回原始字节、词法信息、AST、范围与诊断；`LoadCatalog` 读取 Profile；`Workspace.SetDocument` 支持未保存内容，`Check` 每次重建依赖分析以避免旧缓存；`Format`、`ApplyEdits` 和 `WriteSource` 分别处理格式化、带摘要的字节补丁和受控文件写入。尚无 LSP、GUI 或 Agent 运行时。

## Go 工具验证

```bash
go test ./...
go vet ./...
# Fuzz 名称以测试源码为准，可通过 go test -list Fuzz 列出。
go test -list Fuzz .
```

测试覆盖现有规范/Profile/示例、字段及模块反例、内存覆盖与磁盘变更、宿主边界、精确数值、Unicode 位置、格式化保真、锁漂移及写入冲突。跨平台编译与真实平台运行是独立验收项。

本次 CLI 0.1.0 验证记录（2026-09-22，Go 1.24.5）：

- `go test ./...`、`go vet ./...` 和文档检查通过；解析、宿主提取和格式化的短时 Go fuzz 通过。
- macOS arm64 原生二进制通过完整应用、三种承载、扩展/替换 Profile、标准输入、格式化读写、锁文件、查询命令及退出码冒烟测试；在仓库外的临时目录确认内嵌 Profile 可用。
- Linux amd64/arm64、macOS amd64/arm64、Windows amd64 交叉编译通过。其他平台尚未实际运行验收。

## Python 参考工具环境

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
