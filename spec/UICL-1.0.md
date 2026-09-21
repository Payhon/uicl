<!-- uicl:markdown 1.0 -->

<!-- uicl
document #uicl_standard "UICL 语言标准规范 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# UICL 语言标准规范 1.0

**Universal Intent and Contract Language · 通用意图与契约语言**

语言版本：1.0。Profile 集版本：1.0.0。整合日期：2026-09-20。

这份文件本身采用 UICL 的 Markdown 文档承载格式；同目录 `.md` 是逐字节相同的阅读副本。规范的结构化来源见 `../meta/` 和 `../profiles/`。本版不绑定 AppKernia、Sofvary、具体模型或 Semaquil 产品。

## 阅读导航

第一次阅读建议先看第 1、4、7、9、10、18、22、24 节，再进入所需 Profile。实现者同时阅读 Layout、Compiler、Profile Authoring 和 Conformance。

## 1. 名称、版本与地位

语言名称为 **UICL**，英文全称 **Universal Intent and Contract Language**，中文为“通用意图与契约语言”，标准文件后缀为 `.uicl`。结构化版本头为 `uicl "1.0"`。Semaquil 是一种实现及工具品牌，不是语言、文件格式或唯一解析器。

本文件是前述 AML 0.1/0.2/0.3/1.0、Semaquil 候选、原生文档方案、块式集合、品牌分离以及五类平台场景的 **1.0 整合规范基线**。文档定稿与运行时成熟度分离：本包的 Profile 版本为 `1.0.0`、编辑状态为 `consolidated-draft`，不表示已经完成独立实现互操作认证。

语言、Profile、适配器、模型绑定、工具链和产品版本分别管理。不得因 Semaquil 升级而无条件要求用户修改 UICL 文件。

## 2. 目标与非目标

UICL 统一描述：已有内容、期望结果、允许的处理、输入输出、依赖、验证方式、外部资源和执行回执。它面向内容生成、文件表示、应用开发、Agent、基础设施与训练数据交换。

Core 不负责定义所有平台 API，也不是通用算法语言。无法用已注册领域契约直接描述的算法，可以通过有类型的源码、工具或沙箱组件接入；不能把未定义词汇交给模型临时解释后执行。

“场景完备”指已声明范围内的输入、状态、错误、权限、副作用和恢复有明确处理，不表示模型结果一定正确、所有文件格式都可语义理解、任何平台都已获得支持。

## 3. 规范组成与解释顺序

`spec/UICL-1.0.uicl` 是 Core 与整体语义的阅读规范；`meta/grammar.uicl` 与 `spec/Layout-1.uicl` 是结构语法与布局契约；`profiles/*.uicl` 是字段、主文本槽位、子节点、端口和领域规则的规范源。`guides/`、`.md` 副本、JSON 索引和 EBNF 是派生阅读材料。

MUST=必须，MUST_NOT=禁止，SHOULD=建议，MAY=允许。Core 不可被 Profile 放宽。领域语义可以收紧已允许的行为，但不能改变标点和类型基础规则。同版规范源之间发现冲突须报告 `SPEC_CONFLICT`，不得任意选用有利于执行的一份。

中文 `rule.statement` 是规范性正文，不会仅凭存在就自动编译为验证器。`schema/property/record_schema` 是可由工具读取的结构约束。未实现的领域检查必须报告 `NOT_EVALUATED`。

## 4. 四层结构与多入口

表达层：结构化 UICL、Markdown 承载、HTML 承载、类型化 JSON/YAML 交换。契约层：身份、类型、表达式、引用、依赖、授权、效果和证据。领域层：版本化 Profile。适配层：源码生成、宿主 API、格式导出、数据库、模型和部署工具。

所有受支持入口必须进入相同的类型、权限与效果规则。不能通过换用 JSON、Markdown 或某个适配器，绕过结构化入口禁止的操作。纯文档、纯数据库和纯服务无须包进 app。

## 5. 文件模式与检测

结构化文档以 `uicl "1.0"` 开始，至少含一个节点。Markdown 文档以独立注释 `<!-- uicl:markdown 1.0 -->` 标识；HTML 在 head 中以真实注释 `<!-- uicl:html 1.0 -->` 标识。

可信调用方可显式选择宿主模式，并与文档标记交叉检查。无标记旧文档只能被动导入，不自动激活契约。标记重复、版本不支持、模式冲突或 Core 解析失败必须报告错误，不得自动退回 Markdown。后缀是工具提示，不是执行授权。

## 6. 编码、缩进、标识符与注释

文件采用 UTF-8，可带开头 BOM。读取 LF/CRLF；逻辑值换行按各文本规则处理，但编辑器保留原始字节。无效 Unicode、未配对代理项、NUL 和裸 CR 在结构化输入中拒绝。

结构缩进每级两个空格，不接受 Tab 自动转换，也不允许跳级。原文块内部的 Tab 属于内容。空行和独立 `//` 注释不产生条目。不支持行末注释、分号或一行多个节点。

标识符是 `[A-Za-z_][A-Za-z0-9_]*`，QName 为点连接的标识符。显示内容支持 Unicode。颜色写成双引号字符串；# 只定义身份。中文化编辑器可以使用锁定词典，不允许模型按同义词猜测字段。

## 7. 节点与渐进主文本

完整形式：`种类 [#身份] [主文本] [(属性: 值, ...)]`，后接缩进的属性和子节点。初学者只需“种类、文字、属性、缩进”。行内属性是快捷形式，长配置优先展开。

每个 schema 声明零个或一个 `primary` 槽位。位置主文本必须展开到该普通属性：app/page 为 title，ui.text 为 value，ui.button 为 label，field/input 为 name。image 的主文本是 label，不是 prompt。

同一属性无论出现在块头还是块体，只能出现一次；主文本与其显式属性并存，即使值相同也报 `PRIMARY_CONFLICT`。未声明 primary 的种类不接受主文本。子节点顺序是语义的一部分。

## 8. 节点和值的边界

`theme` 开始节点，`theme:` 开始属性值。Node 体允许属性与子节点；对象值体仅允许键值对；数组体仅允许列表项。首个非空非注释项决定块式值是列表还是记录，不允许混合。

记录里的 `kind: button` 仅是数据，不创建按钮，不带任何执行能力。键名中的点是普通键的一部分，不像某些配置语言那样自动展开嵌套表。JSON/YAML 导入也必须保留这个区别。

## 9. 块式数组与对象

数组既可以写成 `[read, create]`，也可以写成缩进的 `- read`、`- create`。二者产生相同的有序列表。记录可用 `{mode: system}` 或缩进键值对。重复对象键报错；数组元素可以重复。

每个列表 `-` 必须后接空格或换行。`- -10` 表示负数元素。`- name: value` 开始紧凑记录，后续字段与 name 对齐；单独 `-` 后必须有一个更深的非空集合。`[]`、`{}`、`null` 分别表示空列表、空对象、空值；空的 `items:` 报错。

行内 []、{}、() 和表达式均不跨物理行，不接受尾随逗号。长数据使用块式集合。`items: $tasks` 与 `items: [$tasks]` 不等价，后者不自动扁平化。

## 10. 标量、原文与围栏

双引号字符串使用 JSON 风格转义；不提供单引号和模板字符串。简单裸词在值位置是文字，仅 true/false/null 为相应字面量。日期、URI、单位、颜色和币种不自动推断类型。

`|` 文本块去掉固定内容缩进，逻辑换行用 LF；非空文本去掉多余尾部空行并保留一个末尾 LF。空文本使用 `""`。反引号围栏长度至少三，关闭长度与缩进必须匹配；内容含同长关闭符时选择更长围栏。

原文中不解释 @、#、$、// 或节点。源码语言标签不授权运行。需要保留 BOM、非法字节、CRLF 或末尾字节差异时使用 bytes/原始文件资源，而不是依赖文本值。

## 11. 身份、作用域、导入与端口

每个文件是一个模块身份域，所有显式 #id 在该文件内唯一，嵌套不自动产生新的同名模块。匿名节点可以有当前构建内的临时身份，但不能根据行号、内容或序号声称跨编辑稳定。

`module #m` 的 exports 明确导出身份；`use #alias` 的 source 指向模块 URI；`@alias::id` 只能访问导出身份。`@id.port` 必须匹配 schema 声明的端口。相对模块 URI 以声明文件为基址；依赖版本与摘要进入锁文件。

引用图、值依赖图与任务依赖图分别校验。实体关系可以有循环，构建/值求值不得形成无法终止的循环。包内原型保守拒绝循环导入；完整实现可对明确支持的非执行循环另行声明。

## 12. @ 句柄与 $ 数据

@count 是声明/可写句柄，$count 是当前值。只有被 Profile 明确规定的节点才导出运行绑定：state/computed 为页面数据；resource 为查询快照；let/create/update/call/invoke/infer 的有名结果仅在相应顺序作用域之后可见。

动作输入使用 $input，身份使用 $actor，query 使用 $row，policy 使用 $row/$old/$new，组件使用 $props，事件使用 $event。标准 list 模板提供只读 $item；each/for 的 as 提供局部行名。query 行、props 和 actor 不能赋值。保留上下文不可被局部变量遮蔽。

端口引用表示依赖或结果句柄，不自动触发任务，不意味着任务已经执行成功。值的可用阶段须经过检查。

## 13. 类型系统

基础类型为 text、int、decimal、bool、null、unit；领域类型包括 id、date、instant、duration、bytes。组合包括可空类型 T?、list<T>、map<text,T>、记录、枚举和 @命名类型。

int 为有符号 64 位整数；运算溢出必须报错。int 与 decimal 不隐式混算，字符串不自动变成数字/日期，真假值不自动转换。空列表或 null 无法唯一确定业务类型时，必须从属性契约取得类型或显式声明。

形状层的 `value` 表示任意已解析 Core 值，不代表运行时 Any。`record` 表示仍需目标输入契约检查的记录。Profile 中 `dynamic` 输出端口必须根据具体 action/generator 等实例化为确定类型；实例化失败不能发布。

Secret、Spec<T>、Artifact<T>、Generator<I,O>、Task<T>、Slot<T>、ManagedResource、Plan、Evidence 等语义对象不可因 JSON 形状相似而互换。

## 14. 精确数值与文字处理

十进制运行基线以有限值 c×10^e 表示，|c| 的十进制数字不超过 34 位，e 范围为 [-6176, 6111]。零统一为 0。求值结果必须能精确落在该集合，否则返回精度/范围错误，不得暗中转成二进制浮点。该系数/指数形式是对旧稿“decimal128 范围”方向的本次显式冻结，不表示本包已实现数值运行时。

Core 不提供隐式舍入的 /。除法通过 round_div(a,b,scale,mode) 明确小数位和舍入方式，mode 为 half_even、half_up、toward_zero、floor 或 ceiling；零分母报错。% 仅作用于 int，商向零截断。NaN 和 Infinity 不属于 Core。

text 长度以 Unicode 标量数量计，不以 UTF-16 代码单元、字节或可见字形数计。trim 的去除集合固定为 U+0009..000D、0020、0085、00A0、1680、2000..200A、2028、2029、202F、205F、3000，不依赖宿主库当前默认。text 不自动 Unicode 规范化。

## 15. 表达式、纯函数与优先级

只有复合表达式使用 =，单独 $binding 是表达式简写。支持纯函数调用、成员/可空访问、索引、列表/记录构造、比较、布尔与空值合并。运算符从低到高：??、or、and、not、比较/in、+ -、* %、一元 -、访问/调用。?? 右结合，and/or/?? 短路；比较链禁止，必须显式 and。

函数注册表的基线包括 len、trim、str、int、decimal、abs、min、max、sum、contains、starts_with、ends_with、round_div。签名和错误必须固定；无隐式时钟、随机、网络、密钥或 eval。添加纯函数需锁定版本与资源预算，不允许不可信 Profile 自称纯函数后获得副作用权限。

表达式 AST 与源文分别保留。语法解析通过不代表 $ 绑定存在、类型匹配或函数已经实现；这三项是语义编译器的独立检查。

## 16. 资源、产物与受管状态

Resource 包含身份、位置、媒体类型、版本、摘要和访问要求。Artifact 是实际存在、可通过摘要固定的不可变内容。ManagedResource 是外部环境中有稳定身份、可观察状态和管理范围的对象，例如表、云数据库、服务或主机配置。

Spec 表达期望；Plan 描述从观察状态到期望状态的变更；ExecutionRecord 记录真实尝试和结果。文件缺少某个外部对象不表示可以删除它。资源 URI 说明位置，不授予访问权限，哈希说明完整性而非作者身份。

@"./asset.png"、@"file:///C:/Reports/a.pdf"、@"https://example.org/a" 保持统一 URI 入口。appres/artifact 等宿主代理 URI 仅为本项目约定，需要真实 resolver；不声称已注册公共协议。

## 17. 任意文件的三种表示

原生文本承载保留 Python、Go、PHP、TypeScript、JSON、XML、TOML、YAML、Markdown 等源码及其原有语言含义。原始字节承载能够精确表示任何有限文件，包括可执行文件、库、专有二进制和媒体，但不包含文件系统外部元数据。

语义投影由指定适配器产生，声明覆盖范围、已知损失、源摘要和编辑事实来源。没有适配器只报告不透明资源。修改投影必须生成新产物；只有实际字节往返验证才能声称 byte_exact。

不因加载文件定义而加载 DLL、运行源码、连接数据库或向模型发送文件。

## 18. HEXL/1 分行字节协议

HEXL 保持独立版本 1。首行为 !hexl 1 encoding chunk_bytes；数据行为 16 位小写十六进制原始偏移:载荷；尾行为 !end total_bytes line_count sha256:原始字节摘要。支持 hex、base64、base64url、base32。

先按固定原始字节数分块，再逐块独立编码/解码。推荐块长 48，基线上限 1 MiB。非末块必须满长，末块非空，空文件无数据行。偏移从零连续；拒绝重复、缺失、乱序、非法字符、非规范 pad bits、长度或摘要不符、尾部额外非空数据。

base64 使用规范填充；base64url/base32 在本协议使用无填充形式；hex 小写。最终验证前的字节不可信，应写临时文件；验证完成后原子发布。摘要不替代签名。协议不是压缩，也不承诺随机块独立认证。

## 19. Markdown 文档承载

Markdown 基线固定为 CommonMark 0.31.2，其他方言/扩展须明确锁定。标记之后正文按宿主语言阅读；将 .uicl 改为 .md 不修改字节。兼容目标是指定方言的正文静态格式，而不是所有阅读器或字体环境的像素一致。

独立的 <!-- uicl 换行… --> 注解岛保存 Core 节点片段，继承版本和文档模块作用域。只在文档级列 0 的独立块边界启用。Markdown 的 #、-、@、$ 和反引号保持正文意义。普通 aml/uicl 代码围栏只是示例，绝不因此执行。

无标记旧 Markdown 只能经显式 host=markdown 被动导入；不能用任意文本都可为 Markdown 的特点掩盖损坏的结构化源码。

## 20. HTML 与编码载体

HTML 标记位于真实 head 注释中，保留 doctype、charset、原始 HTML/CSS。通过 doc.block target:{html_id:...} 精确绑定已有 ID；缺失或重复 ID 报错。长契约可用 <script type="text/plain" data-uicl>，其载荷是完整结构化文档，不作为 JavaScript。

HTML 注释和 script 都有自己的结束边界。需要时使用 <!-- uicl:base64url ... --> 或 script data-uicl-encoding="base64url"，对完整 UTF-8 载荷一次编码再折行。此算法与 HEXL 的逐块编码不同；必须严格验证字母表、pad bits、大小和 UTF-8。

不可见不是秘密，编码不是加密。原生 HTML 预览和禁脚本的安全预览分别声明；任意脚本可能观察新增节点，服务器 MIME、自身 URL 或第三方清理也可能影响行为，不能承诺绝对透明。

## 21. 正文绑定与无损编辑

next 绑定注解后第一个同级内容块，忽略空白与已识别注解，不进行相似度搜索。正文位置变化导致绑定变化必须显示给作者。HTML ID 按宿主原始字符串匹配，不强行改为 UICL 标识符。

正文是当前发布内容唯一事实源。契约保存意图、约束、资源和变更计划，不保存第二份可独立编辑正文。生成先输出候选与证据，再经批准替换正文。普通阅读器始终显示当前已发布内容。

工具保存原始字节、宿主 CST/AST、Core AST 和双向源码映射。无操作保存字节一致，补丁外保持原样；格式化必须显式选择范围。机器补丁带预期摘要，冲突时重新规划，不覆盖他人修改。

## 22. Profile 定义协议

每个 Profile 必须有 id、version、core、status、summary、requires、origin、implementation、namespaces。每个节点 schema 定义 kind、primary、identity、允许子项、unknown_properties=reject、category、effects、属性与端口。

property 定义形状类型、required、default、expression、enum、targets 与明确语义。record_schema 定义封闭的记录字段。对安全敏感配置不得仅使用未定义对象然后让 Agent 猜规则。generic record/value 的最终契约要由适用业务或适配器补齐，缺失时执行计划不可批准。

Profile 版本精确锁定，依赖无冲突；新增破坏性 primary、默认值、字段含义或执行效果必须升级并迁移。同名标准节点不能被扩展重新解释。扩展不存在不影响被动保存源文，但使用该扩展的主动目标不能执行。

## 23. 标准领域目录与渐进加载

本版目录包含 Meta、Core、Resource、Design、UI、Backend、Application、PostgreSQL、Testing、Deployment、AIGC、Agent、Hybrid、Document、Dataset、Native、Mobile、Cloudflare、Host、Lifecycle 共 20 个 Profile。

标准目录版本固定后，普通作者不需要手写所有 import。工具必须报告实际上启用的 Profile 及适配器能力。目录中出现一个字段不意味着每个实现已经支持它；未知必需能力不能静默忽略。

目录定义了跨领域内容槽位，例如 then 可在测试中包含 expect、UI 容器可包含 if。安装的完整组合必须满足所有实际使用的 schema；不能借助宽泛的子项模式加载未注册节点。平台 .akml/.svml 是兼容 Profile 组合，不另改 Core。

## 24. 渐进应用与组合规则

L0 静态展示，L1 状态事件，L2 collection/form/list，L3 model/query/action/policy/workflow/capability。它们是学习层级，不是四种文件格式。

只有隐式 UI 子项而无 page 的 app 按基础 recipe 生成 home 和 /；显式 page 与隐式首页混用需先展开。collection 必须明确 storage=device/server，默认仅 read/create，不默认 update/delete。server 需要 owner 或完整 policy，以及真实可信身份绑定。

form 自动生成输入、校验、提交状态和去重，list 提供加载、错误、空态和有界分页。默认行为须可展开审查。设备存储改云端存储属于有隐私与迁移影响的显式变更，禁止自动上传旧数据。

## 25. 设计、UI 与跨端一致性

Design 包含 token、主题、样式、断点、语言与可访问性目标。视觉配置不得改变权限、数据类型或业务动作。editable 设计稿需保留指定图层/文字，不能用截图冒充。

UI 是有序声明树。组件 props 为只读、有类型的输入；插槽是结构组合，不是代码注入。ui.input 仅绑定可写 state；on 内动作按顺序等待。表单成功后仅清空仍等于提交快照的输入，刷新失败不能把已成功写入误报为写失败。

跨端首先保证功能、输入输出、错误和交互语义；像素一致作为独立视觉约束。a11y 字段是目标，只有实际验证证据才能证明达标。

## 26. 模型、策略与查询

model 含 id、version、created_at、updated_at 系统字段，作者不能与其重名。新版本默认 version=1，成功更新后原子递增；时间与身份由受信运行时产生，不在纯表达式中读时钟/随机。

模型所有操作默认拒绝。create 检查 new，update 检查 old/new，read 检查 row，delete 检查 old；actor 来自可信服务端。查询先应用策略，再排序、分页、计数和聚合，不能把全量数据交客户端过滤。

query 必须给 select、稳定 order 和 page_size。接口显式暴露，存在 model/action 不自动产生公网 CRUD。身份错误与不可见对象可按安全策略统一报告 NOT_FOUND_OR_FORBIDDEN，避免泄漏存在性。

## 27. 动作、幂等、工作流与异常

action 声明输入、输出、allow、transaction 和 idempotency。非 unit 的所有正常控制路径必须返回匹配类型；run 顺序等待前一步。for 有明确 max_items，调用图不得隐式无界递归。

更新/删除检查 expect_version。事务只包含受管理数据库状态、约束和 outbox，不涵盖外部网络效果。幂等键绑定主体、动作、输入摘要和有效期；相同键不同输入拒绝。外部调用超时可能为 INDETERMINATE，对账前不得无条件重试。

workflow 的状态字段只能经合法 transition 原子更新，转换授权与模型策略同时成立。一般 update 不得绕开状态机。失败路径不能仅靠自然语言“自动恢复”代替明确协议。

## 28. 数据库物理结构与迁移

PostgreSQL Profile 独立于 app/model。pg.table 定义物理列、主键、外键、检查、索引与 RLS；pg.mapping 明确业务字段与系统字段映射。id_conversion=uuid_string 在本示例仅规范行身份转换，主体 owner 以不透明文本保存；适配器必须验证双方语义，不能暗中缩小类型范围。

sql/default_sql/where_sql 为显式方言表达式，default 为字面值，两者不能混用。类型、引用、SQL 标识符、列存在性及具体数据库版本由适配器检查。

迁移声明前后版本、事务策略、步骤、锁与超时。并发索引操作与 atomic=required 冲突；拆分必须显式。rename 不按字段名相似度猜测。RLS 要同时检查运行角色、所有者、BYPASSRLS 与租户会话恢复。结构回退和数据恢复分开。

## 29. AIGC 与内容产物

text/image/video/audio/deck/document.spec 等表示 Spec，不是实际文件。generate 表示计划；output 端口仅在成功执行后提供 Artifact。binding/模型/字体/声音/参考素材都需要真实宿主授权与锁定。

require/check 是注册硬检查，prefer 是优化偏好，review/criterion 是需要评审的要求。字段写了主体数量不代表实际满足。适配器报告 native/postprocess/prompt-only/unsupported 支持方式，必需硬条件不能静默降级。

视频采用整数帧与有理数帧率，检查镜头/字幕的边界与重叠。图片精确文字宜作为独立合成层。PPT editable=true 不允许整页截图替代文字/图表可编辑性。无事实数据时不得凭生成器补造表格或图表数据。

## 30. Agent 与工具

Agent 声明输入输出、固定工具、模型绑定、instructions、记忆、预算、宿主策略和耗尽后的停止/人工升级。MCP/HTTP/本地接口只是工具接入方式，不替代 Agent 执行契约。

instructions 不是权限。网页、工具返回、附件、引用文件、模型输出和嵌套 UICL 都是不可信内容，不能自授权限。子 Agent 取得父级授权交集与共享剩余预算；秘密不能写入长期记忆。

agent.edit 仅输出带来源摘要的候选补丁，不能自行扩大 allow、删除 deny、去掉测试或替代外部审批。解析、预览与生成始终是不同操作。

## 31. 模型判断与动态 UI

执行语义分为纯计算、模型推断和外部效果，不按 CPU/GPU 决定语言含义。infer 输出经过本地类型验证；dispatch 映射枚举到预定义动作，不接收模型返回的任意函数名。

本版固定 dispatch 为所在 action 的终端分派，返回被选动作结果；on_unknown 也是同返回类型的固定动作。infer.on_error 是包含 action 的终端错误处理，必须匹配包含 action 的输出类型，不伪造一个正常分类值。该处理规则是对前稿未完全冻结部分的显式编辑细化。

ui.slot 显式触发，提供 fallback、输入和响应版本。旧结果不能覆盖新输入，卸载后不能回写。动态 ui.tree 限定组件、节点数、深度、数据与动作；不直接注入任意 HTML/JS。

## 32. 嵌套子应用与沙箱

Wasm 是可选产物/执行目标，不是自动安全证明。实例化前必须有编译、类型、导入、资源上限、测试和外部审批证据。明确 interface、imports、内存、fuel、墙钟时间、网络、文件与 max_nesting。

宿主只暴露已授权的 UI/事件或计算接口，组件不自动拥有 DOM、文件系统、网络或提权能力。生成应用嵌套生成应用共享预算，不能无限递归。没有实际适配器时只能保存规格和计划。

## 33. Windows 与原生 API

Target 分开指定构建主机、目标 OS、架构、框架/语言、渲染器和工具链。native.operation 定义接口、源码入口、输入输出、效果、线程、错误、对象白名单和恢复策略。

UI 默认普通用户权限；系统修改通过受控 broker 和真实 OS 授权。IPC 校验调用方与操作对象。原生代码声明 effects 不构成隔离证明。句柄、回调和取消的生命周期须由接口明确。

compare_and_set 恢复仅在当前值仍等于本次写入值时恢复旧值；外部已修改则返回冲突。本包的 C# 夹具故意抛出 NotSupported，不实际改变系统。

## 34. uni-app x 与 React Native

同一业务契约可以投影到多个独立工程。每个 target 固定框架/渲染器、SDK、架构、构建主机、签名和插件。实际共同能力取交集；不支持目标时要求显式分支或失败，不删除功能后假称跨端成功。

mobile.permission 与 mobile.capability 分开描述权限状态、用户手势、拒绝/不可用路径、平台实现和生命周期。声明摄像头/蓝牙等能力不代表用户已授权。后台任务、系统杀进程、页面卸载、取消和重连都应由目标适配器证明对应语义。

本规范定义适配接口，不声称具体版本的 uni-app x/React Native 具有任意目标能力。例子的平台清单只是期望，不是本包实际构建矩阵。

## 35. 受管资源、部署与持久计划

Managed Resource 必须有期望状态、外部身份、管理字段范围、接管方式与移除策略。observe 读取状态；plan 计算差异；apply 在核对计划摘要、观察版本、锁和外部批准后执行。

数据资源默认 retain 或 approval_required，不能因源文删掉节点就删除生产数据库。depends_on 与值/资源绑定形成计划依赖，数组或兄弟节点顺序不等价于串行部署。

每一步记录操作 ID、输入摘要、前置状态、后置状态、失败类别与恢复方式。部分成功保留回执和真实外部 ID；INDETERMINATE 必须先 reconcile。一次运行成功不替代下一次的漂移检查。

## 36. Cloudflare 领域

Provider 区分账号 ID、token 句柄和适配器；资源层分别描述 Worker、R2、D1 和 Hyperdrive。绑定关系形成依赖。兼容日期是作者/锁文件的明确选择，不默认执行当天；账号、配额、区域、套餐和运行时能力由适配器检查。

D1 与 PostgreSQL 不作为同一物理方言互换；需要 PostgreSQL 时显式绑定已有数据库连接方案。Worker 产物不能被当成任意 Linux 进程。

同一资源/字段只有一个管理者。UICL 可以生成 Wrangler/Terraform 等配置，但不能把派生文件和控制台同时作为独立事实来源。重复 apply 复用外部 ID，已有资源需 import；缺失凭据不执行。

## 37. ECS 与通用 SSH 发布

已有 Linux ECS 的机器内部署属于通用 Host Profile。SSH 账户授权与创建 ECS/修改安全组等云控制面权限分离。IP、用户、认证、主机指纹、OS/架构、服务管理器与 sudo 前置条件都要明确。

密码通过 secret/runtime_prompt 或凭据存储供认证适配器使用，不进入命令行、日志、模型上下文或构建缓存。host_key 必须 pinned 并核验可信指纹；变化时报错。sudo 凭据不得默认与 SSH 密码相同。

发布采用临时版本目录、摘要校验、受控切换、健康检查和回执。断线不等于远端失败；重连先对账。previous_binary 只恢复程序版本，数据库无自动回滚。单机重启不宣称零停机。

## 38. 测试、评估与证据

test 使用显式 kind、isolation、given/when/then。test.actor/fixture/reset 仅作用于隔离环境。定位用稳定身份；表单字段由 target+field 标识，不因改显示文字而失效。

结构校验、语义检查、真实行为测试、生成式评估和人工审核是不同证据。状态为 PASS/FAIL/ERROR/NOT_EVALUATED；缺失检查不能视为 PASS。证据绑定被测产物、契约、验证器版本与输入摘要。

本包包含测试声明，但没有执行应用的 E2E/安全测试。tools/tests 的 Python 结果只验证已实现的语法、形状、有限链接和字节封装功能，不能冒充那些业务测试已通过。

## 39. 观测与数据集

日志、指标和追踪声明接收端、保留时间和 payloads=deny；没有授权接收端不上传。凭据、请求正文、完整模型上下文和其他敏感数据不进入默认遥测。

Dataset 记录任务规格、原文/字节/投影、产物、来源组、许可、同意、敏感性和证据。格式开放不表示文件内容可用于训练。unknown/unverified 的权利状态阻断公开训练或再分发。

同源 PDF、截图、译文、Markdown 与投影共享拆分组以防评估泄漏。存档可以文本封装，训练加载应恢复原生模态；性能提升需要实验，不由格式统一直接推导。

## 40. 类型化交换与摘要

规范化节点由 kind、可选 identity、类型化 properties、有序 children 组成。主文本展开为普通属性。每个值带 tag；普通文字 "@logo" 与引用 @logo、原文表达式与已解析 expression AST 都是不同值。

int/decimal 通过无损十进制表示传输，date/instant/bytes/ref 等有独立标签。record 键唯一且不定义业务顺序；list 和 children 保序。JSON/YAML 是这些有类型对象的承载，不是任意文件自动具有执行语义。

sourceDigest、contractDigest、planDigest、artifactDigest、evidenceDigest 分别计算，并包含相关版本域。可变资源在构建前固定真实摘要；模板中的全零/占位摘要不能作为真实证据。本包 SHA256SUMS 仅用于交付文件一致性，不是签名。

## 41. 编译器阶段与一致性类别

处理链：源码字节与宿主模式 → CST/AST → Profile/模块链接 → 名称/类型/纯函数与效果检查 → 确定性 recipe 展开 → Contract Graph → Plan → 外部授权 → 适配器 → 验证/评审 → Artifact/Evidence。

解析器、形状检查器、语义编译器、运行时、代码生成器、编辑器、Agent 适配器分别声明实现类别和 Profile 覆盖。语法合法不等于可构建，可构建不等于已授权，已授权不等于产物通过验收。

解释执行和代码生成须通过同一行为一致性向量。编译/检查不依赖模型在线；Agent 只是契约作者与受控执行参与者。默认展开有版本与来源映射，不由模型临场猜测。

## 42. 错误与编辑协议

诊断包含 code、phase、source span、node/property path、相关位置、说明和可选修复。关键错误包括 SYNTAX、LAYOUT、MISSING_VALUE、DUPLICATE_KEY、PRIMARY_CONFLICT、UNRESOLVED_REF、TYPE_MISMATCH、UNSUPPORTED_CAPABILITY、EFFECT_DENIED、CONFLICT、INDETERMINATE。

编辑器可以建议修复，但不得自动放宽权限、删除测试、关闭主机校验或上传数据来消除错误。对不可信输入的诊断须避免泄漏其他主体资源是否存在。跨模块改名维护 exports/imports，不以简单文本替换作为迁移协议。

## 43. 开放格式、品牌与授权方向

UICL 是厂商中立的格式与规范；Semaquil 是一种实现。任何第三方都可以独立实现并选择自己的产品品牌和代码许可，不需要依赖 Semaquil 账户、模型或云服务。兼容声明必须写明版本与 Profile 范围。

此前讨论的发布方向为：项目有权许可的规范/原创示例采用 CC0-1.0，开源基础代码采用 Apache-2.0，商标、贡献者专利承诺、第三方资料和用户文件权利分别处理。本包记录此发布计划，不替所有贡献者完成权利转让，也不捏造已经建立专利承诺或完成商标注册。

原始参考稿保留来源且不被本包自动重新许可。用户保存为 .uicl 不表示公开、允许训练或将内容归 Semaquil 所有。发布前应由权利人落实许可文本与贡献流程。

## 44. 旧稿迁移与本次新增

历史 AML/Semaquil 文件仍按其原版本解释，不能直接替换头部后当作本版。迁移依次解析旧语法、绑定旧槽位/默认规则、建立旧契约、应用显式转换、生成新文件、比较类型/权限/效果与测试影响。

本版统一 uicl 名称与三种宿主标记；采用 YAML 式块集合；把 provider/SSH/native/PostgreSQL 候选词汇整理为 Profile；正式补入 Managed Resource、Plan、Execution Record 和 Secret 边界。新增依赖边、块体字段形状、绑定范围和分派路径在来源表标注为编辑细化，不冒充前稿全部已有实现。

旧文件与旧测试成绩留作来源证据，不继承为 UICL 新实现的通过结果。

## 45. 本包实现范围与下一步验证

本包提供结构化 Core 语法原型、读取 Profile 的形状检查、主文本展开、包内模块导出与引用身份/端口存在性检查、部分静态反例、宿主示例提取和 HEXL 编解码。实际测试结果以 reports/ 为准。

尚未实现完整运行绑定类型推导、泛型端口兼容性、效果/权限证明、完整规则驱动解析器生成、跨平台 UI/后端/SQL/AIGC/Agent/Wasm 运行时或真实部署器。完整文档无损编辑器也未实现；仅保留原始文件并验证部分字节副本。

正式互操作稳定性应由独立实现与真实平台基准验证。生成模型的图像/文字不要求逐字节相同，但类型、权限、约束处理、证据和错误必须符合相同契约。

## 配套文件与使用示例

[领域 Profile 索引](Profile-Index.uicl) · [Profile 作者指南](Profile-Authoring.uicl) · [编译器与一致性标准](Compiler-Conformance.uicl) · [来源与迁移](Migration-Provenance.uicl)

最小应用：

```uicl
uicl "1.0"

app "你好"
  ui.text "我的第一个 UICL 应用"
```

完整全流程入口：`../examples/fullstack/project.uicl`。生成、部署、原生接口和设备例子都是有边界的契约样例，运行需要真实适配器、输入、凭据、权限和锁文件。
