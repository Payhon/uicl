<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_application "应用组合与渐进默认 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 应用组合与渐进默认 Profile 1.0

标识：`uicl.application@1.0.0`。状态：规范整合稿。依赖：ui, backend。

连接设计、界面、数据和目标平台，简写不隐藏业务决定。


## `app`

应用根；直接 UI 子项可展开为默认 home 页面。

主文本：`title`；身份：`optional`；允许子节点：page, state, computed, collection, ui.*, form, list, resource, component.use。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 应用名称。 |
| `target` | `ref` | 否 | 否 | 单一目标。 |
| `targets` | `list<ref>` | 否 | 否 | 多个独立构建目标，与 target 互斥。 |
| `design` | `ref` | 否 | 否 | 设计系统。 |
| `identity` | `text` | 否 | 否 | 身份由宿主绑定。 |
| `locale` | `text` | 否 | 否 | 默认语言。 |

## `page`

页面与路由。

主文本：`title`；身份：`required`；允许子节点：state, computed, resource, ui.*, form, list, if, each, component.use。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 页面标题。 |
| `route` | `text` | 是 | 否 | 路由。 |
| `access` | `text` | 否 | 否 | 客户端进入要求，不替代后端授权。 |

## `collection`

渐进数据组合；只默认 read/create。

主文本：`None`；身份：`required`；允许子节点：field, policy。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `storage` | `text` | 是 | 否 | 必须显式选择数据位置。 |
| `access` | `text` | 否 | 否 | server 必需的固定授权 recipe。 |
| `operations` | `list<text>` | 否 | 否 | 显式操作集。 |

## 规则 APP_PROGRESSIVE

无 page 的 app 中 UI 子项按基础 recipe 生成 home /，不生成服务器或账号。出现显式 page 时不得再隐式混入另一首页。collection 缺失 storage 不能猜测。


验证类别：`recipe`。

## 规则 APP_STORAGE

server collection 需要可信 identity 和 owner/显式 policy。device 改 server 必须独立取得数据迁移和上传许可。update/delete 绝不因控件出现而自动开放。


验证类别：`recipe-runtime`。
