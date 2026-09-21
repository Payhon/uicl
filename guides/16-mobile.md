<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_mobile "移动端目标和设备能力 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 移动端目标和设备能力 Profile 1.0

标识：`uicl.mobile@1.0.0`。状态：规范整合稿。依赖：core, application。

统一业务契约映射多个工程，权限和生命周期显式。


## `MobileBinding`

平台能力实现。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `ref` | 是 | 否 | 目标平台配置。 |
| `implementation` | `text` | 是 | 否 | 插件/原生模块绑定名。 |

## `mobile.permission`

设备权限状态处理要求。

主文本：`name`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 设备能力名。 |
| `usage` | `text` | 是 | 否 | 前台或后台。 |
| `request` | `text` | 是 | 否 | 请求时机。 |
| `on_denied` | `text` | 是 | 否 | 拒绝路径。 |
| `on_unavailable` | `text` | 是 | 否 | 不支持路径。 |

## `mobile.capability`

跨平台设备接口。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `interface` | `text` | 是 | 否 | 能力接口版本。 |
| `permission` | `ref` | 是 | 否 | 权限契约。 |
| `bindings` | `list<record<MobileBinding>>` | 是 | 否 | 每目标实现。 |
| `input` | `record` | 是 | 否 | 输入类型。 |
| `output` | `text∣ref` | 是 | 否 | 结果类型。 |
| `lifecycle` | `text` | 是 | 否 | 组件销毁后的处理。 |
| `errors` | `list<text>` | 是 | 否 | 权限拒绝/取消/不可用等。 |

输出端口：`output: dynamic`，可用阶段 `execution`。

## 规则 MOBILE_TARGETS

多个 target 分别生成工程，不能把平台能力并集冒充共同能力。未支持的平台/插件报错或要求显式分支。SDK、架构、签名与构建主机锁定。


验证类别：`build-adapter`。

## 规则 MOBILE_PERMISSIONS

声明权限不是用户已授权。处理未询问、允许、拒绝、不可用、取消和后台限制；页面卸载后的回调不能回写失效状态。


验证类别：`device-runtime`。
