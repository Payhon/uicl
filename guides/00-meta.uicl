<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_meta "Profile 与语言自描述 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# Profile 与语言自描述 Profile 1.0

标识：`uicl.meta@1.0.0`。状态：规范整合稿。依赖：无。

定义 Profile、节点形状、记录形状、规则、文法与示例。自描述不是无需受信引导器。


## `profile`

领域注册单元；版本、依赖及节点语义必须显式。

主文本：`title`；身份：`required`；允许子节点：schema, record_schema, rule。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `id` | `text` | 是 | 否 | 可解析的唯一 Profile 标识，不代表已经注册域名。 |
| `version` | `text` | 是 | 否 | 精确的领域版本。 |
| `core` | `text` | 是 | 否 | 所需 Core 版本。 |
| `status` | `text` | 是 | 否 | 规范编辑状态，不是运行时认证。 |
| `title` | `text` | 是 | 否 | 显示名称。 |
| `summary` | `text` | 是 | 否 | 领域的目标与排除范围。 |
| `requires` | `list<text>` | 是 | 否 | 依赖的 Profile 精确标识与版本。 |
| `origin` | `text` | 是 | 否 | 来源或本次细化标识。 |
| `implementation` | `text` | 是 | 否 | 本包仅结构检查，执行适配器的实现状态。 |
| `namespaces` | `list<text>` | 是 | 否 | 所占用的节点前缀；不能改变 Core。 |

## `schema`

一个节点的结构契约，未知属性默认拒绝。

主文本：`kind`；身份：`required`；允许子节点：property, port。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `kind` | `text` | 是 | 否 | 被定义的 QName。 |
| `summary` | `text` | 是 | 否 | 节点的含义。 |
| `primary` | `text∣null` | 是 | 否 | 唯一主文本槽位；null 表示不接受。 |
| `identity` | `text` | 是 | 否 | 身份要求。 |
| `children` | `list<text>` | 是 | 否 | 允许的子节点种类或已注册的命名空间模式。 |
| `unknown_properties` | `text` | 是 | 否 | 基线固定 reject。 |
| `effects` | `list<text>` | 是 | 否 | 可能出现的副作用分类，不等于授权。 |
| `category` | `text` | 是 | 否 | 领域对象类别。 |

## `record_schema`

具名记录值的结构约束，不创建节点或权限。

主文本：`name`；身份：`required`；允许子节点：property。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 记录类型名。 |
| `summary` | `text` | 是 | 否 | 字段集合语义。 |
| `unknown_fields` | `text` | 是 | 否 | 禁止未知字段。 |

## `property`

一个属性的类型、必需性、默认值与表达式边界。

主文本：`name`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 属性键。 |
| `type` | `text` | 是 | 否 | 形状类型表达式。 |
| `description` | `text` | 是 | 否 | 字段语义与边界。 |
| `required` | `bool` | 是 | 否 | 是否必须出现。 |
| `expression` | `bool` | 是 | 否 | 允许延后到业务类型检查的纯表达式。 |
| `default` | `value` | 否 | 否 | 缺失时采用的确定性默认值；校验工具不自行注入。 |
| `enum` | `list<value>` | 否 | 否 | 可接受的字面量集合。 |
| `targets` | `list<text>` | 否 | 否 | 引用允许的节点种类；端口另按端口类型检查。 |

## `port`

节点可引用的输出端口声明。

主文本：`name`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 端口名称。 |
| `type` | `text` | 是 | 否 | 业务输出类型。 |
| `phase` | `text` | 是 | 否 | 编译时或执行后可用。 |

## `rule`

规范性规则及验证类别；正文不会自动执行。

主文本：`code`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `code` | `text` | 是 | 否 | 跨版本稳定的规则标识。 |
| `level` | `text` | 是 | 否 | 规范强度。 |
| `statement` | `text` | 是 | 否 | 完整规范正文。 |
| `verification` | `text` | 是 | 否 | 所需验证种类。 |
| `origin` | `text` | 是 | 否 | 历史基线、场景讨论或本次编辑细化。 |

## `grammar`

结构化文法入口。

主文本：`title`；身份：`required`；允许子节点：g.rule。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 文法名称。 |
| `start` | `ref` | 是 | 否 | 起始产生式。 |
| `input` | `text` | 是 | 否 | token/布局契约说明。 |
| `status` | `text` | 是 | 否 | 规范状态。 |

## `g.rule`

命名产生式。

主文本：`name`；身份：`required`；允许子节点：g.seq, g.choice, g.repeat, g.ref, g.token, g.literal。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 规则名。 |

## `g.seq`

按顺序匹配全部子项。

主文本：`None`；身份：`optional`；允许子节点：g.seq, g.choice, g.repeat, g.ref, g.token, g.literal。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `g.choice`

按声明顺序尝试；不能吞掉已确认的语义错误。

主文本：`None`；身份：`optional`；允许子节点：g.seq, g.choice, g.repeat, g.ref, g.token, g.literal。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `g.repeat`

在指定上下界重复一个子模式。

主文本：`None`；身份：`optional`；允许子节点：g.seq, g.choice, g.ref, g.token, g.literal。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `min` | `int` | 是 | 否 | 最少次数。 |
| `max` | `int∣null` | 是 | 否 | 最多次数，null 为未指定语法上界；仍受实现资源限制。 |

## `g.ref`

引用本模块的文法规则。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `ref` | 是 | 否 | 目标产生式。 |

## `g.token`

匹配词法器提供的 token。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | token 种类。 |

## `g.literal`

匹配固定文法文本。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `text` | `text` | 是 | 否 | 固定字面量。 |

## `example`

显式记录可提取示例，不在读取阶段运行。

主文本：`title`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 用例说明。 |
| `path` | `text` | 是 | 否 | 包内相对源码位置。 |
| `profiles` | `list<text>` | 是 | 否 | 涉及领域。 |
| `readiness` | `text` | 是 | 否 | 执行准备状态。 |
| `body` | `text` | 否 | 否 | 可选内嵌原文，必须与独立文件一致。 |

## `document`

阅读文档元数据，不创建运行应用。

主文本：`title`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 文档标题。 |
| `version` | `text` | 是 | 否 | 规范版本。 |
| `status` | `text` | 是 | 否 | 发布状态。 |
| `authority` | `text` | 是 | 否 | 权限或标准机构声明边界。 |

## 规则 PROFILE_SELF_DESCRIPTION

Profile 是惰性数据。schema 可以限定本 Profile 的词汇，不能覆盖宿主的授权、Core 运算符、密钥隔离或其他已锁定 Profile。同名 schema/记录/依赖版本冲突必须拒绝。


验证类别：`static`。

## 规则 PROFILE_BOOTSTRAP

元规范仍需受信的最小解析器；本包检查器读取 schema/property/record_schema，文法 g.* 是自描述与可派生数据，未作为完整解析器自动生成依据。


验证类别：`tooling`。
