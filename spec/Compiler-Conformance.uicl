<!-- uicl:markdown 1.0 -->

<!-- uicl
document #compiler "UICL 编译器、运行时与一致性标准"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# UICL 编译器、运行时与一致性标准

## 1. 规范处理阶段

源码保真读取 → 宿主/Core 解析 → Profile 和模块链接 → 形状与 primary 检查 → 运行绑定/业务类型检查 → 纯函数与效果检查 → 确定性 recipe 展开 → 依赖和目标能力检查 → 生成 Plan → 外部授权 → 执行 → 验证产物与证据。

每个阶段输出带源位置的结果；未执行的阶段不能报告通过。解析器不会调用模型补齐语法。编译器不是部署授权者。

## 2. 本包工具

`tools/syntax.py`：结构语法原型，支持节点、块式与行内集合、原文、引用、表达式 AST。它不执行表达式，也不实现完整业务类型系统。

`tools/check.py`：从 profiles/*.uicl 实际加载 schema、property 和 record_schema；检查 primary、未知/缺失属性、封闭记录、子节点、模块 exports、引用身份/端口存在性及部分静态反例。表达式结果类型、秘密流向与泛型端口类型仍需完整编译器。

`tools/hexl.py`：独立流式编码/解码及临时文件验证发布。此工具保留先前协议，不运行解码后的内容。

`tools/host.py`：宿主示例检查适配器。Markdown 利用已安装的 MarkdownIt CommonMark 模式提取真实 html_block；HTML 采用受限示例解析器，并明确不是完整 WHATWG 树构建一致性实现。代码示例不自动执行。

## 3. 一致性层级

| 层级 | 必须证明 |
|---|---|
| Syntax | 结构、值、布局与表达式语法，原文/字符串边界 |
| Shape | schema、必填、primary、子节点与封闭记录 |
| Link | 模块依赖、exports、身份与端口 |
| Semantic | 作用域、类型、纯函数、效果、权限与默认展开 |
| Adapter | 目标 API、SQL、构建和文件保真 |
| Runtime | 状态、并发、幂等、错误与恢复 |
| Host editing | 宿主兼容、注解定位、无损补丁与冲突 |

本包只覆盖前三层的列明子集，以及独立 HEXL 测试和有限宿主例子。不能用多个低层测试通过推导高层已实现。

## 4. 执行计划的最低内容

计划包含 contractDigest、inputsDigest、观察到的资源版本、适配器锁、资源管理范围、依赖图、逐步前置/后置条件、幂等标识、超时、恢复类别和权限需求。批准绑定计划摘要与有效期。

每步的外部 ID 与执行回执持久保存。PENDING/RUNNING/SUCCEEDED/FAILED/CANCELLED/INDETERMINATE 的转换显式记录，断线不把 RUNNING 直接当 FAILED。状态存储中的秘密必须引用或脱敏。

## 5. 必须有的反例

重复属性、错误 primary、字符串伪引用、缺失模块导出、未知端口、数组扁平化误解、空属性、列表缩进错误、代码围栏伪注解、宿主模式冲突、未授权能力、无效证据、并发索引与强制事务、错误 SSH 主机指纹、上传损坏、断线重试和旧响应覆盖。

测试 .uicl 只描述测试，本包 Python suite 执行的是语言工具与载荷验证，不是 Windows/移动端/PostgreSQL/Cloudflare/ECS 的真实任务。

## 6. 自描述的边界

meta/grammar.uicl 是真实结构化产生式，generated/core.ebnf 从它派生。tools/syntax.py 是独立手写引导器，尚未根据文法自动生成或独立验证所有 token 流。schema/property 已实际驱动形状检查。改变 schema 会改变测试结果；改变 rule.statement 不会神奇生成业务检查代码。

自描述、有限自校验、编译器自举与形式化正确性证明是四个不同概念。
