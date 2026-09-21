<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_document "原生文档与结构化文档规格 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 原生文档与结构化文档规格 Profile 1.0

标识：`uicl.document@1.0.0`。状态：规范整合稿。依赖：core, resource, testing。

Markdown/HTML 正文是发布内容事实源，注解附加契约。


## `doc.block`

绑定原生正文，不复制第二份正文。

主文本：`None`；身份：`required`；允许子节点：require, review。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `text∣record` | 是 | 否 | next 或 {html_id:...}，由宿主解析。 |
| `intent` | `text` | 否 | 否 | 编辑/内容意图，不当作硬规则。 |
| `spec` | `ref` | 否 | 否 | 可选关联的生成规格。 |
| `expected_digest` | `text` | 否 | 否 | 补丁前正文版本。 |

## `document.spec`

文档生成规格，可要求导出可编辑文档。

主文本：`label`；身份：`required`；允许子节点：doc.section, require, review。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `label` | `text` | 否 | 否 | 标题。 |
| `language` | `text` | 是 | 否 | 语言。 |
| `output` | `text` | 是 | 否 | 媒体类型。 |
| `editable` | `bool` | 是 | 否 | 保留语义结构的要求。 |

## `doc.section`

具有标题和原生正文的文档部分。

主文本：`title`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 章节名。 |
| `body` | `text` | 是 | 否 | Markdown 或纯文正文。 |
| `format` | `text` | 否 | 否 | 正文方言。 |

## 规则 HOST_PARSE_FIRST

先由宿主解析真实注释/数据块；代码围栏、script/style/textarea/template 或转义文本中的标记不激活。Markdown 注解仅文档级列零独立块。错误不能静默降级为 Markdown。


验证类别：`host-parser`。

## 规则 HOST_PRESERVATION

无操作保存必须字节不变，补丁外字节保持。正文为唯一可编辑事实源；注解只存规格/检查/来源。改后缀承诺锁定宿主下静态可读，不承诺所有渲染器像素或任意脚本行为不变。


验证类别：`host-editor`。

## 规则 HOST_BOUNDARY

注释/script 的终止边界不能靠 UICL 字符串引号保护；复杂载荷使用严格 Base64url 编码。编码不是加密。next 绑定不到对象时报错；机器编辑检查旧摘要。


验证类别：`host-parser`。
