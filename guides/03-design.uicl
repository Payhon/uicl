<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_design "设计系统与设计稿 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 设计系统与设计稿 Profile 1.0

标识：`uicl.design@1.0.0`。状态：规范整合稿。依赖：resource。

设计 token、主题、样式、响应式、多语言与可编辑画布。


## `Breakpoint`

响应式断点。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 断点名。 |
| `min_width` | `int` | 是 | 否 | 最小逻辑宽度。 |
| `columns` | `int` | 是 | 否 | 布局列数。 |

## `design`

可复用设计系统。

主文本：`label`；身份：`required`；允许子节点：theme, style, a11y。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 系统名称。 |
| `preset` | `text` | 否 | 否 | 可选锁定预设。 |
| `mode` | `text` | 否 | 否 | 主题模式。 |
| `tokens` | `record` | 是 | 否 | 分类的 token 值；类型由设计适配器校验。 |
| `locales` | `list<text>` | 否 | 否 | 支持语言。 |
| `breakpoints` | `list<record<Breakpoint>>` | 否 | 否 | 有序响应式规则。 |

输出端口：`tokens: DesignTokens`，可用阶段 `compile`。

## `theme`

对基础 token 的主题覆盖。

主文本：`name`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 主题名。 |
| `tokens` | `record` | 是 | 否 | 覆盖相同 token 类型，不改变权限。 |

## `style`

可复用视觉属性集合。

主文本：`name`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 样式名。 |
| `values` | `record` | 是 | 允许，仍需业务检查 | 布局与视觉属性，禁止业务副作用。 |

## `a11y`

可访问性目标，不自动构成达标证据。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `text` | 是 | 否 | 验收基线名称。 |
| `reduce_motion` | `text` | 否 | 否 | 减少动画偏好。 |
| `keyboard` | `bool` | 否 | 否 | 要求键盘可操作。 |

## `design.canvas`

可编辑设计画布。

主文本：`label`；身份：`required`；允许子节点：design.layer。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 设计稿标题。 |
| `size` | `list<int>` | 是 | 否 | 二维尺寸，必须两个正整数。 |
| `editable` | `bool` | 是 | 否 | 要求对象可编辑。 |

## `design.layer`

带几何和语义的独立图层。

主文本：`label`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 图层说明。 |
| `kind` | `text` | 是 | 否 | 图层种类。 |
| `bounds` | `list<number>` | 是 | 否 | [x,y,w,h]，单位由画布固定。 |
| `text` | `text` | 否 | 否 | 文字层正文。 |
| `source` | `ref` | 否 | 否 | 图片层资源。 |
| `fill` | `text` | 否 | 否 | 颜色或 token 名。 |

## 规则 DESIGN_NO_BUSINESS_EFFECT

主题、断点和样式只能改变布局或视觉；不得改变动作授权、字段类型或后端行为。token 类型冲突报错。


验证类别：`type-effect`。

## 规则 DESIGN_EDITABILITY

editable 要求输出保留指定的文字/对象/图层。截图不能冒充可编辑设计稿；a11y 目标必须经独立检查或评审。


验证类别：`artifact-validation`。
