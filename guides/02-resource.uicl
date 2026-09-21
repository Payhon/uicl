<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_resource "文件、产物与任意字节 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 文件、产物与任意字节 Profile 1.0

标识：`uicl.resource@1.0.0`。状态：规范整合稿。依赖：core。

原文、字节与语义投影分别建模，引用不隐式执行。


## `file`

现有文件的受控引用或内嵌原文。

主文本：`label`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 可读名称。 |
| `source` | `ref` | 否 | 否 | 源 URI。 |
| `media` | `text` | 是 | 否 | 媒体类型。 |
| `body` | `text` | 否 | 否 | 内嵌原文，与 source 互斥。 |
| `digest` | `text` | 否 | 否 | 原始字节摘要；实际值不能伪造。 |
| `verification` | `text` | 否 | 否 | 需由构建锁固定真实摘要。 |
| `encoding` | `text` | 否 | 否 | 文本编码说明，不隐式转码。 |

## `bytes`

内嵌 HEXL 或已有字节资源。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `media` | `text` | 是 | 否 | 媒体类型。 |
| `transport` | `text` | 是 | 否 | 独立字节封装协议。 |
| `body` | `text` | 是 | 否 | HEXL 原文。 |

## `projection`

绑定原始字节的语义投影。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `source` | `ref` | 是 | 否 | 来源文件。 |
| `adapter` | `text` | 是 | 否 | 投影适配器。 |
| `fidelity` | `text` | 是 | 否 | 保真声明。 |
| `editable` | `bool` | 是 | 否 | 是否允许修改投影。 |
| `known_loss` | `list<text>` | 是 | 否 | 已知损失，未知不能写无损。 |
| `source_digest` | `text` | 否 | 否 | 执行时核验的来源摘要。 |

输出端口：`output: Projection`，可用阶段 `execution`。

## `artifact`

已经存在或外部交付的不可变产物。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `media` | `text` | 是 | 否 | 产物媒体类型。 |
| `source` | `ref` | 是 | 否 | 产物 URI。 |
| `digest` | `text` | 否 | 否 | 真实摘要或执行绑定前置条件。 |
| `producer` | `text` | 否 | 否 | 来源执行回执标识。 |

## 规则 RESOURCE_BYTES

原始字节可以承载任何有限文件，但不包括权限、目录关系、扩展属性等文件系统元数据。没有格式适配器只报告不透明资源；不能声称理解内部结构。


验证类别：`codec-adapter`。

## 规则 RESOURCE_URI

相对 URI 以声明模块为基准。解析/编辑不打开任意资源；受控解析限制路径、协议、重定向、大小和递归。内容摘要不等于来源签名。


验证类别：`resolver`。

## 规则 RESOURCE_PROJECTION

投影记录源摘要、适配器版本、损失和编辑事实来源。修改投影生成新产物；byte_exact 只能由字节往返验证证明。


验证类别：`adapter`。
