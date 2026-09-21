# AML 1.0：自描述规范整合包

**Abstract Mark Language，统一意图、资源与执行契约语言。**

语言版本为 `aml "1.0"`，规范版本为 `1.0.0`，状态为 `consolidated-proposal`。这是 0.1、0.2、0.3 的原创整合提案，不是任何标准组织认证、互操作成熟度或生产可用性的声明。

## 从哪里读

直接阅读 `AML-1.0.aml`。它是约 240 KiB 的单文件阅读版，包含真实 AML 节点描述的语言规则、文法、节点/属性注册表和可提取用例，不是把整篇 Markdown 藏进一个字符串。

维护和开发时，以 `spec/*.aml`、`profiles/*.aml` 及用例目录为入口。**用例正文的维护源是 `spec/14-examples.aml`，`examples/*.aml` 是供阅读与执行检查的提取副本**。修改用例正文后通过 `rebuild.py` 同步，不同时编辑两份。主文件也由 AML 分模块源重建，不双向同时编辑。

| 入口 | 内容 |
|---|---|
| `spec/00-overview.aml` | 版本、定位、原始来源摘要、自描述边界 |
| `spec/01-grammar.aml` | 33 条真正的 AML 文法树产生式 |
| `spec/02-lexical.aml` | 编码、缩进、字符串、原文块、符号与侧条件 |
| `spec/03-types-expressions.aml` | 类型、纯表达式、精度、绑定、作用域 |
| `spec/04-resources.aml` | 统一引用、原文字节与语义投影 |
| `spec/05-progressive.aml` | 隐式页面、状态推断、collection/form/list |
| `spec/06-application.aml` | 权限、查询、动作、事务、幂等、UI、工作流、HTTP |
| `spec/07-content.aml` | 文本、图片、视频、音频、PPT、文档、设计 |
| `spec/08-agent-hybrid.aml` | Agent、推断、动态 UI、沙箱子应用 |
| `spec/09-lifecycle-dataset.aml` | 设计、测试、发布、观测和训练交换 |
| `spec/10-compiler-runtime.aml` | 统一契约图、适配器、锁与摘要 |
| `spec/11-hexl.aml` | 独立 HEXL/1 协议 |
| `spec/12-meta-selfdescription.aml` | 元规范：如何解释文法和 schema |
| `spec/13-conformance.aml` | 当前实现范围与未实现项 |
| `spec/14-examples.aml` | 35 个可提取独立 AML 用例 |
| `spec/15-migration.aml`、`16-integration-additions.aml` | 显式记录版本冲突和整合取舍 |
| `spec/17-domain-clarifications.aml` | 1.0 明确补充的领域接口子集 |
| `spec/18-source-coverage.aml` | 前三版各章节的覆盖映射 |
| `profiles/*.aml` | 6 组元/资源/应用/内容/Agent/生命周期注册表，198 种节点形状 |
| `tests/conformance-vectors.aml` | 53 个合法与非法 AML 测试向量 |

## 三行入口

```aml
aml "1.0"
app #hello "你好"
  ui.text "我的第一个应用"
```

语法只要求掌握节点、两个空格缩进、`key: value`。需要引用时才使用 `#id`、`@ref`、`$value`；复杂表达式前使用 `=`。原文使用 `|` 或反引号围栏。

198 个注册节点是整个跨领域词汇库，不是新手的必背清单。`g.*`、`schema` 和 `rule` 是规范编写者使用的 Meta Profile，普通应用无需书写。

## 怎样自描述

`g.rule`、`g.seq`、`g.choice`、`g.repeat`、`g.ref`、`g.token` 和 `g.literal` 构成结构化文法树。`aml_meta.py` 实际从 AML 加载这些节点来识别 token 流。

`schema` 和 `property` 定义节点属性、必填、类型形状、标签、锚点、子节点与未知属性规则。`schema` 自己也有一个 `schema` 定义，注册表会接受同样的形状检查。

`rule` 记录规范等级、适用阶段、来源、验证方式和中文规范正文。中文规则不是通用可执行代码；完整类型、权限、预算和运行语义必须由相应实现验证，当前报告把这些要求保留为 `NOT_EVALUATED`。

`example.body` 保存独立 AML 文档；自检工具会提取、重新解析、检查，并确认它与 `example.path` 的独立文件逐字一致。内层锚点不会泄漏到外层规范，示例在加载时也不会执行。

仍然需要一个受信的最小引导解析器。这里实现的是**自描述与有限自校验**，不是“完全不需要外部程序”、不是编译器由 AML 编写，也不是数学上的正确性证明。

## 运行参考工具

需要 Python 3.10 或更高版本，只用标准库，没有第三方安装依赖。解压后进入 `aml-1.0` 目录。

```bash
# 解析单个文档，生成语法 AST。
python reference/aml_seed.py examples/02-counter.aml --out generated/counter.syntax.json

# 从 AML 分模块规范重建主文件、提取用例、派生 EBNF。
python reference/rebuild.py

# 解析规范、加载 AML 文法和 schema、复核用例与测试向量。
python reference/selfcheck.py --out reports/selfcheck.json

# 验证语言工具和 HEXL 编解码器。
python -m unittest discover -s tests -v

# 流式二进制文本封装，不执行载荷。
python reference/hexl.py pack assets/sample.bin generated/sample.b64l --encoding base64url --chunk 48
python reference/hexl.py unpack generated/sample.b64l generated/restored.bin
```

HEXL 默认拒绝覆盖已有目标；只有明确使用 `--force` 才允许替换。库函数 `unpack` 校验完成前写出的字节是临时、不可信字节；CLI 使用临时文件验证后发布。

本包没有提供已实现的 `aml dev`、`aml build`、`aml generate` 或 `aml deploy` 命令。上述 `aml_seed.py` 只生成 AST，不运行应用。

## 已实际验证什么

本次交付的实际结果记录于 `reports/selfcheck.json` 和 `reports/unittest.txt`：

- 62 个 AML 文档通过已实现的语法、数据驱动 token 文法、节点形状及有限锚点检查。
- 35 个内嵌 AML 用例与其独立文件一致。
- 53 个合法或非法向量符合预期接受/拒绝结果。
- 57 个 Python 测试方法通过，包含 HEXL 的 360 组编码/块长/文件大小组合往返子测试。
- 8 个原始来源快照的 SHA-256 与清单一致。
- 134 条规范要求的完整语义状态仍为 `NOT_EVALUATED`，不能用前述通过结果代替。

测试还验证了：修改 AML 中的文法字面量会改变文法识别结果；修改 AML 中的属性类型会改变结构校验结果。因此文法与 schema 不是只展示给人看的装饰。

## 尚未实现和严格限制

当前没有完整业务类型系统、跨模块符号链接、纯函数类型验证、全部端口和数据绑定范围检查、权限/效果分析器、全部默认 recipe 展开、UI/数据库/AIGC/Agent/Wasm 运行时、业务测试执行器、实际部署或完整规范 IR 编解码器。宽泛的结构属性 `value` 和子节点 `*` 不代表运行时可以接受任意语义。

规范里的表达式没有交给 Python/JavaScript eval；参考工具不会根据 URI 自动联网、打开文件、安装 Profile、加载 DLL 或运行围栏代码。外部资源和部分端口被明确记录为 deferred。

用例中的示例 URI、模型/字体/设计预设、MCP 工具和 Wasm 接口不代表现实中存在已注册资源。生成、部署、设备和服务端例子需要真正的宿主绑定、权限与预算；缺失时必须阻止执行，不应补一个假成功。

## 版本整合依据

0.1 的应用执行语义是保留基线；0.2 的渐进默认与生命周期被保留；0.3 的通用资源、符号、AIGC、Agent 和混合执行体系为统一表层。0.x 源码不能只改版本头后直接当作 1.0。

主要整合取舍包括 `#` 身份、`@` 引用、`$` 值；`ui.text`/文本生成规格分离；`ui.dispatch`/业务分类 `dispatch` 分离；Agent 宿主策略 `agent_policy` 与数据 `policy` 分离；精确除法继续 `round_div`，`not` 优先级沿用 0.1；逐项来源和需要审阅的变更已写成 AML。

新增元规范与部分领域细化标为 `1.0-self-description`、`1.0-integration-decision` 或 `1.0-domain-completion`，不冒充旧稿已经确定的内容。

`HEXL` 保持其独立版本 1。`.akml` 和 `.svml` 只是兼容 Profile 组合规划，不改变 AML Core；此包未实现两种平台方言。

## 文件维护

`generated/core.ebnf` 是由 AML 文法派生的人类阅读副本，不应反过来作为第二份手工规范。主文件中的相对资源 URI 已按原模块地址重定位，内嵌示例的 URI 仍按其提取位置解释。

`sources/` 保留原稿，是历史证据，不是 1.0 的另一套执行规则。`SHA256SUMS.txt` 用于本包文件一致性检查，不是作者身份签名或可信发布认证。
