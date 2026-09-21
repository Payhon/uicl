<!-- uicl:markdown 1.0 -->

<!-- uicl
document #readme "UICL 1.0 规范、Profile 与示例包"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目规范整合稿；工具范围以实际验证报告为准。"
-->

# UICL 1.0 规范、Profile 与示例包

**Universal Intent and Contract Language，通用意图与契约语言。** 标准后缀 `.uicl`。Semaquil 是一种实现品牌，不是格式或读取文件的唯一入口。

本包整合前面讨论中的渐进语法、块式集合、Markdown/HTML 原生文档、AIGC、Agent、全生命周期以及 Windows、移动端、PostgreSQL、Cloudflare、ECS/SSH 场景。1.0 表示规范整合版本，不表示所有适配器或运行时已经实现。

## 从哪里开始

[语言标准主文档](spec/UICL-1.0.uicl) 是 Markdown 承载的 UICL 文件，可直接阅读，或只改后缀为 `.md`。包内已经提供[字节相同的 Markdown 副本](spec/UICL-1.0.md)。

[结构化单文件全集](UICL-1.0-Structured.uicl) 包含 Core 规则、结构化文法、20 个 Profile 与内嵌领域示例。它是从规范源派生的阅读/机器分析版，不要独立修改。

[快速入门](spec/Quickstart.uicl)、[Profile 索引](spec/Profile-Index.uicl)、[Profile 作者指南](spec/Profile-Authoring.uicl)、[全流程示例说明](spec/Fullstack-Walkthrough.uicl)、[编译器与一致性](spec/Compiler-Conformance.uicl)、[编辑细化](spec/Domain-Semantics.uicl)、[来源与迁移](spec/Migration-Provenance.uicl)。

## 文件布局

```text
spec/       主规范、布局、HEXL、编译器、Profile 作者指南、迁移
meta/       Core 规则镜像、自描述文法、带源码的示例目录
profiles/   20 个机器可读领域定义，字段/记录/端口/规则
examples/   42 份 UICL 用例，包括 9 模块的完整应用契约
  fullstack/         需求、设计、后端、数据库、UI、测试、构建、部署、项目
  hosted/            Markdown/HTML 示例及后缀副本
  profile-authoring/ 第三方 Profile 及消费该 Profile 的示例
guides/     每个 Profile 的字段与语义阅读指南
assets/     少量安全示例资源与 HEXL 载荷，不含生产可执行程序
provenance/ 历史原稿快照与摘要，非当前规范源
reports/    本轮实际检查，不复用旧版本通过结论
tools/      语法原型、形状/引用检查、有限宿主检查、HEXL、视图重建
tests/      语言工具和 HEXL 的 Python 自动化测试
generated/ 派生的 EBNF、文法树和用例索引
```

规范有 45 节，20 个 Profile 合计定义 162 种节点和 20 种封闭记录；Core 45 条加领域 51 条，共 96 条规则。节点词汇是整个领域目录，不是新手需要背诵的列表。

## 运行工具

结构、Profile 与 HEXL 工具使用 Python 3.10+ 标准库；Markdown 宿主检查另需 `requirements-host.txt` 指定的 markdown-it-py。此依赖版本记录当前验证环境，不表示宣称它是最新版本。

```bash
python -m pip install -r requirements-host.txt

# 解析，不运行应用。
python tools/syntax.py examples/02-counter.uicl --out generated/counter.syntax.json

# 标准 Profile、例子、跨模块引用和有限静态检查。
python tools/check.py --out reports/structural-check.json
python tools/check.py examples/fullstack/project.uicl

# 显式加载第三方 Profile，不自动从网络安装。
python tools/check.py examples/profile-authoring/measurement-example.uicl \
  --extra-profile examples/profile-authoring/measurement-profile.uicl

# Markdown/HTML 示例提取与形状检查。
python tools/host.py --out reports/host-check.json

# 从规范源重建文法阅读视图、内嵌示例和结构化单文件全集。
python tools/rebuild_views.py
python tools/rebuild_views.py --check

# 全套可重复检查；写 reports/verification.json。
python tools/verify_package.py

# 单独运行自动化测试。
python -m unittest discover -s tests -v

# HEXL 只解码文件，不执行内容；默认拒绝覆盖。
python tools/hexl.py pack assets/sample.bin generated/sample.b64l --encoding base64url --chunk 48
python tools/hexl.py unpack generated/sample.b64l generated/restored.bin
```

## 全流程示例怎样衔接

入口是 `examples/fullstack/project.uicl`。它把业务需求、设计系统、Go 后端契约、PostgreSQL 物理结构、Web UI、验收测试、构建包和 SSH/systemd 部署计划连接起来。所有文件有明确 exports，引用通过模块别名解析；数据库迁移先于应用发布的关系使用依赖边，不依靠数组排列猜测顺序。

示例只请求连接已有 Linux 主机，不代替申请云资源。登录密码、主机指纹、签名和云 Token 由宿主提供，缺少真实绑定不得执行。Windows 示例中的 C# 仅是显式抛出未实现错误的占位接口；不修改任何系统服务。移动、模型和沙箱示例也不是现成插件安装包。

## 已验证与未验证

本轮自动化测试执行 106 个测试方法，其中 HEXL 往返矩阵包含 360 组子测试。实际报告还分别列出结构文档、宿主文档、第三方 Profile、内嵌示例和来源摘要检查。

这些测试不证明完整 UICL 语义已经实现。完整运行绑定类型推导、泛型端口兼容、纯函数与权限/效果证明、全部 recipe 展开、真实 UI/后端/数据库/AIGC/Agent/Wasm/部署运行时都未实现；相关领域规则完整验证状态为 `NOT_EVALUATED`。HTML 提取器仅支持普通良构示例，不是完整 WHATWG 树构建器；未运行本轮浏览器像素或平台真机测试。只读源码保留不等于已实现无损编辑器。

本包没有真实 `uicl dev/build/deploy` 命令。工具不会解析密码值、自动联网导入、调用模型或启动外部进程部署。SQL 语义、凭据策略和业务执行不能由形状检查代替。

## 维护与开放边界

主规范以 `spec/UICL-1.0.uicl` 为阅读规范源，机器字段定义以 `profiles/*.uicl` 为源，文法以 `meta/grammar.uicl` 为源，独立用例以 `examples/` 为源。镜像、单文件书和索引不可单独成为另一份手工事实来源。同版源文件冲突应报告 SPEC_CONFLICT。

格式保持厂商中立，可由不同实现和平台 Profile 使用。前面约定的 CC0 规范、Apache-2.0 代码及单独专利/品牌规则属于正式发布前需要落实的授权方案；本包不代替所有权利人完成授权，也不重许可 provenance/ 中的第三方资料。用户把内容保存成 UICL 不等于同意公开、训练或转让文件内容权利。

`uicl.lock.json` 锁定本包 Profile 的实际摘要与工具验证环境。适配器/凭据仍标为 UNBOUND，它不是生产部署锁或外部服务可用性保证。`SHA256SUMS.txt` 用于包内容一致性检查，不是可信签名或认证。
