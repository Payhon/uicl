<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_native "桌面原生 API 与系统操作 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 桌面原生 API 与系统操作 Profile 1.0

标识：`uicl.native@1.0.0`。状态：规范整合稿。依赖：core, resource。

原生 API 通过受限类型边界、进程、ABI 和恢复契约接入。


## `NativePolicy`

允许系统对象与权限。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `allowed_objects` | `list<text>` | 是 | 否 | 允许操作的对象白名单。 |
| `execution` | `text` | 是 | 否 | 执行进程位置。 |
| `approval` | `text` | 是 | 否 | 每次系统写操作须外部批准。 |

## `NativeRecovery`

系统修改恢复条件。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `capture_before` | `bool` | 是 | 否 | 先保存旧值。 |
| `restore` | `text` | 是 | 否 | 仅当前值仍为本次写入值时恢复。 |

## `native.operation`

有 ABI/权限边界的原生能力。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `interface` | `text` | 是 | 否 | 版本化系统操作接口。 |
| `implementation` | `ref` | 是 | 否 | 源码文件。 |
| `entry` | `text` | 是 | 否 | 已知入口。 |
| `input` | `record` | 是 | 否 | 输入类型。 |
| `output` | `text∣ref` | 是 | 否 | 结果类型。 |
| `effects` | `list<text>` | 是 | 否 | 系统副作用。 |
| `policy` | `record<NativePolicy>` | 是 | 否 | 白名单与执行权限。 |
| `recovery` | `record<NativeRecovery>` | 是 | 否 | 恢复。 |
| `threading` | `text` | 是 | 否 | 线程/单元模型要求。 |
| `errors` | `list<text>` | 是 | 否 | 失败分类。 |

输出端口：`output: dynamic`，可用阶段 `execution`。

## 规则 NATIVE_BROKER

界面按普通权限运行；提权由受控 broker 和操作系统授权完成。IPC 验证调用方和输入白名单。effects 是声明，不是任意原生代码的隔离证明。


验证类别：`os-adapter`。

## 规则 NATIVE_LIFETIME

ABI、架构、线程、句柄所有权、回调、取消与错误映射必须在接口中固定。恢复 compare-and-set，外部已修改时返回 CONFLICT，不盲目覆盖。


验证类别：`os-adapter`。
