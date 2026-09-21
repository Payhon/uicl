<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_hybrid "推断、生成槽位与沙箱子应用 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 推断、生成槽位与沙箱子应用 Profile 1.0

标识：`uicl.hybrid@1.0.0`。状态：规范整合稿。依赖：core, ui, backend, aigc, agent。

模型判断与确定性分派分离，动态区域有边界。


## `generator`

可复用的有类型生成能力。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `input` | `record` | 是 | 否 | 输入形状。 |
| `output` | `text∣ref` | 是 | 否 | Artifact 或命名类型。 |
| `binding` | `text` | 是 | 否 | 实现绑定。 |
| `instructions` | `text` | 是 | 否 | 生成意图。 |
| `limits` | `record<CallLimits>` | 是 | 否 | 预算。 |
| `constraints` | `record` | 是 | 否 | 由输出 Profile 校验的约束记录。 |
| `sandbox` | `record` | 否 | 否 | 需要代码执行时，完整隔离要求。 |

输出端口：`output: dynamic`，可用阶段 `execution`。

## `ui.slot`

可装入生成产物的动态区域。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `generator` | `ref` | 是 | 否 | 生成器。 |
| `args` | `record` | 是 | 允许，仍需业务检查 | 输入。 |
| `trigger` | `text` | 是 | 否 | 显式或用户事件触发。 |
| `fallback` | `ref` | 是 | 否 | 普通阅读时可用的既有资源。 |
| `response_policy` | `text` | 否 | 否 | 响应版本策略。 |

## `infer`

模型推断节点，只把结果作为数据。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `using` | `ref` | 是 | 否 | 有类型能力/Agent/生成器。 |
| `args` | `record` | 是 | 允许，仍需业务检查 | 输入。 |
| `on_error` | `ref` | 是 | 否 | 固定失败动作。 |

输出端口：`result: dynamic`，可用阶段 `execution`。

## `dispatch`

固定白名单分派。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `value` | `value` | 是 | 允许，仍需业务检查 | 枚举推断值。 |
| `routes` | `record` | 是 | 否 | 枚举文字到已声明动作的映射。 |
| `on_unknown` | `ref` | 是 | 否 | 未识别值/非法输出路径。 |

## `wasm.sandbox`

Wasm 实例化契约，不是任意源码的安全保证。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `artifact` | `ref` | 是 | 否 | 已编译验证的 Wasm 产物。 |
| `interface` | `text` | 是 | 否 | 宿主接口版本。 |
| `imports` | `list<text>` | 是 | 否 | 允许导入白名单。 |
| `network` | `text` | 是 | 否 | 网络权限。 |
| `filesystem` | `text` | 是 | 否 | 文件权限。 |
| `memory_mib` | `int` | 是 | 否 | 内存上限。 |
| `fuel` | `int` | 是 | 否 | 执行步数预算。 |
| `timeout_ms` | `int` | 是 | 否 | 墙钟超时。 |
| `max_nesting` | `int` | 是 | 否 | 嵌套深度。 |
| `approval` | `text` | 是 | 否 | 实例化外部授权。 |

## 规则 HYBRID_INFER

模型只提供通过本地类型检查的数据。dispatch 映射固定且必须覆盖枚举或显式 unknown；不得接受模型给出的任意函数/命令。timeout/refusal/非法值进入错误路径。


验证类别：`type-runtime`。

## 规则 HYBRID_UI

动态 UI 限定允许组件、最大节点/深度、可绑定数据与动作。Wasm 必须经过编译、类型、导入与测试检查；宿主仍负责权限和资源限制。


验证类别：`adapter-runtime`。
