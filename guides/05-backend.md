<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_backend "后端业务与 API Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 后端业务与 API Profile 1.0

标识：`uicl.backend@1.0.0`。状态：规范整合稿。依赖：core。

模型、访问策略、查询、动作、并发、幂等和工作流。


## `model`

服务端实体；保留 id/version/created_at/updated_at 系统字段。

主文本：`name`；身份：`required`；允许子节点：field, policy。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 实体显示名。 |

## `policy`

服务端行级访问策略。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `read` | `bool` | 否 | 允许，仍需业务检查 | row 上的读取授权。 |
| `create` | `bool` | 否 | 允许，仍需业务检查 | new 上的创建授权。 |
| `update` | `bool` | 否 | 允许，仍需业务检查 | old/new 上的更新授权。 |
| `delete` | `bool` | 否 | 允许，仍需业务检查 | old 上的删除授权。 |

## `query`

有界只读查询。

主文本：`None`；身份：`required`；允许子节点：input。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `model` | `ref` | 是 | 否 | 实体模型。 |
| `allow` | `bool` | 是 | 允许，仍需业务检查 | 调用授权。 |
| `where` | `bool` | 否 | 允许，仍需业务检查 | 额外行过滤。 |
| `select` | `list<text>` | 是 | 否 | 明确输出字段。 |
| `order` | `list<text>` | 是 | 否 | 稳定排序，分页必须包含唯一 tiebreaker。 |
| `page_size` | `int` | 是 | 否 | 单页大小上限。 |

输出端口：`result: QueryPage`，可用阶段 `execution`。

## `input`

动作或查询的命名输入。

主文本：`name`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 输入名。 |
| `type` | `text∣ref` | 是 | 否 | 业务类型。 |
| `normalize` | `text` | 否 | 否 | trim 或 none。 |
| `min` | `int` | 否 | 否 | 长度/数值下界。 |
| `max` | `int` | 否 | 否 | 上界。 |

## `action`

确定性业务控制流；模型推断须显式节点。

主文本：`name`；身份：`required`；允许子节点：input, run。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 动作说明。 |
| `returns` | `text∣ref` | 是 | 否 | 输出业务类型。 |
| `allow` | `bool` | 是 | 允许，仍需业务检查 | 服务端调用授权。 |
| `transaction` | `text` | 是 | 否 | 数据事务要求。 |
| `idempotency` | `text` | 是 | 否 | 服务端去重约定。 |

输出端口：`result: dynamic`，可用阶段 `execution`。

## `run`

顺序执行块。

主文本：`None`；身份：`optional`；允许子节点：let, create, update, delete, return, throw, assert, if, for, call, invoke, emit, infer, dispatch, transition.apply。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `let`

局部不可变绑定，只在当前 run/事件之后可见。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `value` | `value` | 是 | 允许，仍需业务检查 | 表达式。 |

## `create`

通过模型策略及约束的原子数据create

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `model` | `ref` | 是 | 否 | 模型。 |
| `values` | `record` | 是 | 允许，仍需业务检查 | 写入字段，系统字段由宿主管理。 |

输出端口：`result: dynamic`，可用阶段 `execution`。

## `update`

通过模型策略及约束的原子数据update

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `model` | `ref` | 是 | 否 | 模型。 |
| `values` | `record` | 否 | 允许，仍需业务检查 | 写入字段，系统字段由宿主管理。 |
| `id` | `value` | 是 | 允许，仍需业务检查 | 记录身份。 |
| `expect_version` | `int` | 是 | 允许，仍需业务检查 | 乐观锁版本。 |

输出端口：`result: dynamic`，可用阶段 `execution`。

## `delete`

通过模型策略及约束的原子数据delete

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `model` | `ref` | 是 | 否 | 模型。 |
| `values` | `record` | 否 | 允许，仍需业务检查 | 写入字段，系统字段由宿主管理。 |
| `id` | `value` | 是 | 允许，仍需业务检查 | 记录身份。 |
| `expect_version` | `int` | 是 | 允许，仍需业务检查 | 乐观锁版本。 |

输出端口：`result: dynamic`，可用阶段 `execution`。

## `return`

从动作返回有类型结果。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `value` | `value` | 是 | 允许，仍需业务检查 | 结果。 |

## `throw`

显式业务错误。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `code` | `text` | 是 | 否 | 稳定错误码。 |
| `message` | `text` | 否 | 否 | 脱敏错误说明。 |

## `assert`

违反前置条件时中止，不能替代身份授权。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `test` | `bool` | 是 | 允许，仍需业务检查 | 布尔条件。 |
| `code` | `text` | 是 | 否 | 失败码。 |

## `if`

二分条件，条件必须 bool。

主文本：`None`；身份：`optional`；允许子节点：then, else。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `test` | `bool` | 是 | 允许，仍需业务检查 | 纯表达式。 |

## `then`

条件分支。

主文本：`None`；身份：`optional`；允许子节点：ui.*, let, set, call, invoke, create, update, delete, return, throw, assert, if, for, refresh, notify, expect。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `else`

条件分支。

主文本：`None`；身份：`optional`；允许子节点：ui.*, let, set, call, invoke, create, update, delete, return, throw, assert, if, for, refresh, notify。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `for`

有界服务端循环。

主文本：`None`；身份：`optional`；允许子节点：run。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `items` | `list<value>` | 是 | 允许，仍需业务检查 | 输入列表。 |
| `as` | `text` | 是 | 否 | 循环绑定名。 |
| `max_items` | `int` | 是 | 否 | 强制迭代上限。 |

## `each`

有稳定 key 的 UI 列表循环。

主文本：`None`；身份：`optional`；允许子节点：ui.*, if。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `items` | `list<value>` | 是 | 允许，仍需业务检查 | 输入集合。 |
| `as` | `text` | 是 | 否 | 行绑定名。 |
| `key` | `value` | 是 | 允许，仍需业务检查 | 每项稳定唯一 key。 |

## `call`

调用固定动作，不接受模型生成的函数名。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `action` | `ref` | 是 | 否 | 动作声明。 |
| `args` | `record` | 是 | 允许，仍需业务检查 | 输入参数。 |

输出端口：`result: dynamic`，可用阶段 `execution`。

## `emit`

事务 outbox 事件；不在事务内直接调用网络。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `event` | `text` | 是 | 否 | 版本化事件类型。 |
| `payload` | `record` | 是 | 允许，仍需业务检查 | 无秘密的事件数据。 |

## `endpoint`

显式 HTTP 接口映射。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `method` | `text` | 是 | 否 | HTTP 方法。 |
| `path` | `text` | 是 | 否 | 相对 API 路径。 |
| `action` | `ref` | 否 | 否 | 处理动作。 |
| `query` | `ref` | 否 | 否 | 读取查询。 |
| `auth` | `text` | 是 | 否 | 公共或身份验证。 |
| `body` | `text` | 否 | 否 | 请求体编码。 |
| `rate_limit` | `record` | 否 | 否 | 由 HTTP 适配器验证的限流参数。 |

## `service`

后端服务聚合入口。

主文本：`name`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 名称。 |
| `target` | `ref` | 是 | 否 | 编译目标。 |
| `models` | `list<ref>` | 是 | 否 | 包含的模型。 |
| `endpoints` | `list<ref>` | 是 | 否 | 显式暴露接口。 |
| `identity_binding` | `text` | 是 | 否 | 外部可信身份实现。 |
| `health_path` | `text` | 否 | 否 | 无敏感内容健康检查端点。 |

## `workflow`

持久有限状态机。

主文本：`None`；身份：`required`；允许子节点：transition。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `model` | `ref` | 是 | 否 | 关联模型。 |
| `state_field` | `text` | 是 | 否 | 被保护的状态字段。 |
| `initial` | `text` | 是 | 否 | 初始状态。 |
| `terminal` | `list<text>` | 是 | 否 | 终态。 |

## `transition`

只能在已授权且版本吻合时改变状态。

主文本：`name`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 迁移名。 |
| `from` | `text` | 是 | 否 | 源状态。 |
| `to` | `text` | 是 | 否 | 目标状态。 |
| `allow` | `bool` | 是 | 允许，仍需业务检查 | 转换授权，与模型策略同时成立。 |

## `transition.apply`

原子应用工作流转换。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `transition` | `ref` | 是 | 否 | 已声明转换。 |
| `id` | `value` | 是 | 允许，仍需业务检查 | 记录身份。 |
| `expect_version` | `int` | 是 | 允许，仍需业务检查 | 期望版本。 |

## 规则 BACKEND_AUTH

服务端权限默认拒绝。模型策略在分页/计数/聚合前应用；API 或按钮存在不自动授予权限。actor 来自可信身份适配器，不能从请求体声明。


验证类别：`runtime`。

## 规则 BACKEND_ATOMICITY

事务只涵盖受管理数据、并发约束与 outbox。事务内禁止直接网络副作用。system fields 由运行时生成，id 为不透明身份，version 初始 1 且每次成功更新递增。


验证类别：`runtime`。

## 规则 BACKEND_IDEMPOTENCY

去重键绑定主体、动作、有效期和输入摘要。相同键不同输入报错。外部调用结果不确定必须进入 INDETERMINATE，对账前不得重复执行。


验证类别：`runtime`。

## 规则 BACKEND_FLOW

动作非 unit 的每个正常路径必须返回匹配类型；调用图不得隐式无界递归。有界循环、if、return/throw 语义固定。普通 update 不能绕过 workflow 改变状态字段。


验证类别：`type-effect-runtime`。
