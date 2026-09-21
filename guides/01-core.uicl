<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_core "核心类型、模块与执行上下文 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 核心类型、模块与执行上下文 Profile 1.0

标识：`uicl.core@1.0.0`。状态：规范整合稿。依赖：meta。

模块、参数、状态、能力与目标环境；不绑定具体产品。


## `CallLimits`

执行预算。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `timeout_ms` | `int` | 是 | 否 | 墙钟超时，正整数。 |
| `attempts` | `int` | 是 | 否 | 最大尝试次数。 |
| `max_steps` | `int` | 否 | 否 | 最大推理/控制步数。 |
| `max_tool_calls` | `int` | 否 | 否 | 最大工具调用次数。 |
| `max_cost_minor` | `int` | 否 | 否 | 最小货币单位预算。 |
| `currency` | `text` | 否 | 否 | 预算币种；有费用上限时必需。 |

## `module`

当前文件的导出边界。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 模块显示名。 |
| `exports` | `list<text>` | 是 | 否 | 导出的本文件显式身份名称。 |

## `use`

导入模块；解析工具仅在获准的包根内读取。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `source` | `ref` | 是 | 否 | 本地或受控的模块 URI。 |
| `version` | `text` | 否 | 否 | 可选精确模块版本。 |

## `parameter`

由宿主提供的普通输入，不是秘密。

主文本：`name`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 参数说明。 |
| `type` | `text∣ref` | 是 | 否 | 输入业务类型。 |
| `default` | `value` | 否 | 否 | 明确的缺省输入。 |
| `required` | `bool` | 否 | 否 | 未提供时阻断执行。 |

## `secret`

受控凭据句柄。禁止 value/default 与秘密正文。

主文本：`name`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 凭据用途。 |
| `source` | `text` | 是 | 否 | 凭据解析方式。 |
| `binding` | `text` | 否 | 否 | 不含密钥的宿主查找标识。 |
| `scope` | `list<text>` | 是 | 否 | 允许消费该句柄的适配器标识。 |

## `type`

命名记录或枚举业务类型。

主文本：`name`；身份：`required`；允许子节点：field。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 类型说明。 |
| `enum` | `list<text>` | 否 | 否 | 固定枚举值。 |

## `field`

模型/记录/表单的命名字段。

主文本：`name`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 字段名称。 |
| `label` | `text` | 否 | 否 | 显示名称。 |
| `type` | `text∣ref` | 否 | 否 | 业务类型；缺省 text。 |
| `min` | `int∣decimal` | 否 | 否 | 长度或数值下界。 |
| `max` | `int∣decimal` | 否 | 否 | 长度或数值上界。 |
| `default` | `value` | 否 | 否 | 纯字面缺省值。 |
| `enum` | `list<value>` | 否 | 否 | 有限取值。 |
| `immutable` | `bool` | 否 | 否 | 创建后不可修改。 |
| `normalize` | `text` | 否 | 否 | 输入前处理。 |

## `state`

当前组件/页面的可写状态。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `value` | `value` | 是 | 允许，仍需业务检查 | 初值。 |
| `type` | `text∣ref` | 否 | 否 | 无法唯一推断时必须给定。 |

## `computed`

只读派生数据，不得参与赋值。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `value` | `value` | 是 | 允许，仍需业务检查 | 纯表达式。 |
| `type` | `text∣ref` | 否 | 否 | 期望输出类型。 |

## `target`

生成与运行目标，不表示已安装工具链。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `platform` | `text` | 否 | 否 | 单一目标平台。 |
| `platforms` | `list<text>` | 否 | 否 | 分别生成的平台清单。 |
| `framework` | `text` | 否 | 否 | 目标框架。 |
| `language` | `text` | 否 | 否 | 源码语言。 |
| `adapter` | `text` | 是 | 否 | 宿主适配器绑定名。 |
| `architecture` | `text` | 否 | 否 | 目标架构。 |
| `architectures` | `list<text>` | 否 | 否 | 分别构建的架构。 |
| `toolchain` | `text` | 是 | 否 | 精确工具链由锁文件固定。 |
| `build_host` | `text` | 否 | 否 | 构建主机要求。 |
| `ui_privilege` | `text` | 否 | 否 | Windows UI 执行级别。 |
| `renderer` | `text` | 否 | 否 | 目标渲染器。 |

## `capability`

有类型的外部/原生能力。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `interface` | `text` | 是 | 否 | 锁定的接口契约。 |
| `binding` | `text` | 是 | 否 | 实现提供者。 |
| `input` | `record` | 是 | 否 | 字段名到业务类型的记录；调用前按类型契约检查。 |
| `output` | `text∣ref` | 是 | 否 | 结果业务类型。 |
| `effects` | `list<text>` | 是 | 否 | 副作用集合。 |
| `permissions` | `list<text>` | 是 | 否 | 请求的宿主能力。 |
| `limits` | `record<CallLimits>` | 是 | 否 | 超时与重试上限。 |
| `errors` | `list<text>` | 是 | 否 | 接口声明的失败码。 |
| `trust` | `text` | 否 | 否 | 结果可信度。 |

输出端口：`output: dynamic`，可用阶段 `execution`。

## `invoke`

调用已声明能力，执行时获取输入并等待。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `capability` | `ref` | 是 | 否 | 目标能力。 |
| `args` | `record` | 是 | 允许，仍需业务检查 | 按目标输入类型检查，不允许 shell 字符串拼接。 |
| `on_error` | `ref` | 否 | 否 | 已声明错误处理动作。 |

输出端口：`output: dynamic`，可用阶段 `execution`。

## 规则 CORE_PURE

表达式只能使用显式 $ 绑定或锁定的纯函数。不得读密钥、网络、时钟、随机源、数据库写入或任意 eval。未知函数不是自动调用模型的请求。


验证类别：`type-effect`。

## 规则 CORE_REF_SCOPE

每个文件模块 #id 唯一；@alias::id 需要 use 与目标 exports。@id.port 必须为目标声明的端口。标题不参与身份。$ 绑定由上下文提供，不是任意锚点自动导出的值。


验证类别：`static-and-type`。

## 规则 CORE_SECRET

Secret 只能进入声明接收凭据的适配器入口；不能转为文字、进入表达式/提示词/训练样本/日志/普通状态文件。解析不能读取 source 指向的密钥。


验证类别：`type-effect-runtime`。

## 规则 CORE_TARGET

构建主机、目标框架、目标平台、运行位置与权限分别检查。多个 target 表示多个产物；缺失目标能力不得静默删功能。


验证类别：`adapter`。
