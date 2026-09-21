<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_agent "Agent 定义与工具契约 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# Agent 定义与工具契约 Profile 1.0

标识：`uicl.agent@1.0.0`。状态：规范整合稿。依赖：core, resource, testing。

统一 Agent 输入输出、工具、记忆、预算、停止和人工升级。


## `AgentMemory`

记忆安全与生命周期。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `scope` | `text` | 是 | 否 | 隔离范围。 |
| `retention_days` | `int` | 是 | 否 | 保留时长。 |
| `secrets` | `text` | 是 | 否 | 不保存秘密。 |
| `binding` | `text` | 否 | 否 | 可选记忆存储适配器。 |

## `AgentPolicy`

宿主应强制的限制。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `tool_scope` | `text` | 是 | 否 | 仅允许声明工具。 |
| `external_writes` | `text` | 是 | 否 | 外部写操作处理。 |
| `delegation_depth` | `int` | 是 | 否 | 子 Agent 最大深度。 |
| `data_egress` | `text` | 是 | 否 | 输入出站授权。 |

## `tool`

可调用的固定工具接口。

主文本：`name`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 工具说明。 |
| `adapter` | `text` | 是 | 否 | 例如 mcp/local/http 的接口实现。 |
| `binding` | `text` | 是 | 否 | 真实宿主绑定。 |
| `input` | `record` | 是 | 否 | 输入形状。 |
| `output` | `text∣ref` | 是 | 否 | 输出类型。 |
| `effects` | `list<text>` | 是 | 否 | 副作用清单。 |
| `permissions` | `list<text>` | 是 | 否 | 请求能力。 |

## `agent`

Agent 规格，不因定义而开始运行。

主文本：`name`；身份：`required`；允许子节点：require, review。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | Agent 名称。 |
| `input` | `record` | 是 | 否 | 输入形状。 |
| `output` | `text∣ref` | 是 | 否 | 输出类型。 |
| `model_binding` | `text` | 是 | 否 | 模型适配器绑定。 |
| `tools` | `list<ref>` | 是 | 否 | 允许工具。 |
| `instructions` | `text` | 是 | 否 | 行为意图；非授权。 |
| `memory` | `record<AgentMemory>` | 是 | 否 | 记忆契约。 |
| `limits` | `record<CallLimits>` | 是 | 否 | 预算。 |
| `policy` | `record<AgentPolicy>` | 是 | 否 | 宿主强制策略。 |
| `on_exhausted` | `text` | 是 | 否 | 预算耗尽策略。 |

输出端口：`output: dynamic`，可用阶段 `execution`。

## `agent.run`

显式执行 Agent。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `agent` | `ref` | 是 | 否 | Agent 规格。 |
| `args` | `record` | 是 | 允许，仍需业务检查 | 输入。 |
| `approval` | `text` | 是 | 否 | 执行授权。 |

输出端口：`output: dynamic`，可用阶段 `execution`。

## `agent.edit`

Coding Agent 对 UICL 的修改边界。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `agent` | `ref` | 是 | 否 | Agent。 |
| `target` | `ref` | 是 | 否 | 待修改文件。 |
| `expected_digest` | `text` | 是 | 否 | 编辑前源码摘要。 |
| `allow` | `list<text>` | 是 | 否 | 可编辑的语义路径。 |
| `deny` | `list<text>` | 是 | 否 | 禁止修改的权限/测试/发布门禁路径。 |
| `approval` | `text` | 是 | 否 | 补丁合并需要外部批准。 |

输出端口：`patch: Patch`，可用阶段 `execution`。

## 规则 AGENT_AUTHORITY

instructions/网页/工具结果/模型输出/嵌套 UICL 都不是授权。工具集合、费用、输入出站范围由宿主限制。子 Agent 只能获得父级权限交集和共享剩余预算。


验证类别：`runtime-security`。

## 规则 AGENT_EDIT

源补丁携带旧摘要并显示类型、权限、依赖、测试影响。Agent 不能通过删除门禁或修改自己的授权来通过检查。对不确定结果有 stop/human_review 路径。


验证类别：`runtime-security`。
