<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_host "SSH 主机操作与 ECS 发布 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# SSH 主机操作与 ECS 发布 Profile 1.0

标识：`uicl.host@1.0.0`。状态：规范整合稿。依赖：core, deployment。

通用已有 Linux 主机部署。SSH 不授予云控制面权限。


## `SshAuthentication`

SSH 认证。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `method` | `text` | 是 | 否 | 认证方式。 |
| `credential` | `ref` | 是 | 否 | 秘密句柄。 |

## `HostKey`

可信服务器身份。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `policy` | `text` | 是 | 否 | 固定指纹。 |
| `fingerprint` | `text` | 是 | 允许，仍需业务检查 | 可信渠道取得的主机指纹。 |

## `HostExpected`

主机前置检查。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `os` | `text` | 是 | 否 | 操作系统。 |
| `architecture` | `text` | 是 | 否 | CPU 架构。 |
| `service_manager` | `text` | 是 | 否 | 服务管理器。 |

## `Privilege`

远程操作提权要求。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `method` | `text` | 是 | 否 | sudo 或 none。 |
| `mode` | `text` | 是 | 否 | 提权方式。 |
| `credential` | `ref` | 否 | 否 | 单独 sudo 凭据，不默认等于 SSH 密码。 |

## `Health`

健康检查。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `url` | `text` | 是 | 否 | 目标 URL；受执行位置和网络策略限制。 |
| `expected_status` | `int` | 是 | 否 | 期望 HTTP 状态。 |
| `timeout_ms` | `int` | 是 | 否 | 超时。 |

## `ReleaseRecovery`

发布恢复。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `on_unhealthy` | `text` | 是 | 否 | 程序版本恢复范围。 |
| `on_disconnect` | `text` | 是 | 否 | 断线先对账。 |
| `database` | `text` | 是 | 否 | 不自动逆转数据库。 |

## `ops.ssh_host`

已有主机的可信连接配置。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `address` | `text` | 是 | 允许，仍需业务检查 | IP 或主机名。 |
| `port` | `int` | 否 | 否 | SSH 端口。 |
| `user` | `text` | 是 | 允许，仍需业务检查 | SSH 用户。 |
| `authentication` | `record<SshAuthentication>` | 是 | 否 | 认证。 |
| `host_key` | `record<HostKey>` | 是 | 否 | 服务端身份。 |
| `expected` | `record<HostExpected>` | 是 | 否 | 前置环境。 |
| `privilege` | `record<Privilege>` | 是 | 否 | 提权。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

## `ops.service_release`

上传校验、版本目录、进程管理及健康检查。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `host` | `ref` | 是 | 否 | SSH 主机。 |
| `artifact` | `ref` | 是 | 否 | 真实构建产物。 |
| `service_name` | `text` | 是 | 否 | 服务名。 |
| `run_as` | `text` | 是 | 否 | 独立服务账户，须预先存在或经引导创建。 |
| `directory` | `text` | 是 | 否 | 受管发布目录。 |
| `executable` | `text` | 是 | 否 | 包内可执行入口。 |
| `environment` | `record` | 否 | 否 | 环境变量；秘密只在运行端定向注入。 |
| `retain_releases` | `int` | 是 | 否 | 保留程序版本数量。 |
| `health` | `record<Health>` | 是 | 否 | 健康检查。 |
| `recovery` | `record<ReleaseRecovery>` | 是 | 否 | 失败/断线恢复。 |
| `approval` | `text` | 是 | 否 | 发布需要外部授权。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

## 规则 HOST_AUTH

SSH 密码/密钥仅由凭据代理交给认证入口。严格主机指纹校验，变化时报错；不能为方便而自动关闭验证。sudo 权限必须独立检查。


验证类别：`ssh-adapter`。

## 规则 HOST_RECONCILE

断线不能直接解释为失败。执行日志/标识与目标状态决定恢复；先对账后重试迁移。上传到临时版本目录、校验摘要后切换；previous_binary 不回滚数据库。


验证类别：`ssh-runtime`。

## 规则 HOST_CLOUD_BOUNDARY

登录 ECS 的系统权限与阿里云 RAM/安全组/实例创建权限分离。普通 SSH 场景不需要也不拥有云控制面权限；网络和账户前置条件不满足应阻断。


验证类别：`adapter-preflight`。
