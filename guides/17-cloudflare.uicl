<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_cloudflare "Cloudflare 受管基础设施 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# Cloudflare 受管基础设施 Profile 1.0

标识：`uicl.cloudflare@1.0.0`。状态：规范整合稿。依赖：core, deployment。

提供方专属资源绑定，生命周期服从 Deployment 契约。


## `cloudflare.provider`

Cloudflare 账号与凭据句柄。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `account_id` | `text` | 是 | 允许，仍需业务检查 | 账号 ID。 |
| `token` | `ref` | 是 | 否 | 最小权限 token 句柄。 |
| `adapter_binding` | `text` | 是 | 否 | 具体 API/Terraform 等适配器。 |

## `cloudflare.r2_bucket`

明确保留策略的受管云资源。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `provider` | `ref` | 是 | 否 | 账号提供方。 |
| `name` | `text` | 是 | 否 | 云端名称。 |
| `on_remove` | `text` | 是 | 否 | 数据资源删除策略。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

输出端口：`resource: CloudResource`，可用阶段 `execution`。

## `cloudflare.d1_database`

明确保留策略的受管云资源。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `provider` | `ref` | 是 | 否 | 账号提供方。 |
| `name` | `text` | 是 | 否 | 云端名称。 |
| `on_remove` | `text` | 是 | 否 | 数据资源删除策略。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

输出端口：`resource: CloudResource`，可用阶段 `execution`。

## `cloudflare.worker`

Worker 部署单元与绑定。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `provider` | `ref` | 是 | 否 | 账号提供方。 |
| `name` | `text` | 是 | 否 | Worker 名称。 |
| `entry` | `ref` | 是 | 否 | 已构建 JS/模块产物。 |
| `compatibility_date` | `text` | 是 | 否 | 锁定兼容日期，不自动取今天。 |
| `workers_dev` | `bool` | 是 | 否 | 是否公开 workers.dev 入口。 |
| `bindings` | `record` | 是 | 否 | 绑定名到受管资源/凭据句柄，按接口验证。 |
| `on_remove` | `text` | 否 | 否 | 删除门禁。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

## `cloudflare.hyperdrive`

连接已有 PostgreSQL 等支持源的连接资源。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `provider` | `ref` | 是 | 否 | 账号提供方。 |
| `name` | `text` | 是 | 否 | 名称。 |
| `origin` | `ref` | 是 | 否 | 连接凭据句柄。 |
| `on_remove` | `text` | 是 | 否 | 移除策略。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

## 规则 CF_DIALECT

D1 使用独立 SQL 方言，不把 PostgreSQL RLS/类型/索引静默投影为 D1。需要 PostgreSQL 时声明外部数据库与可用连接方案。


验证类别：`adapter`。

## 规则 CF_MANAGED_STATE

通过已锁定适配器管理每项资源。重复 apply 复用外部 ID，要求 import 接管、state 锁与漂移检测。同字段不能同时由多个管理工具独立覆盖。


验证类别：`provider-runtime`。
