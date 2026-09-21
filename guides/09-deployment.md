<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_deployment "构建、受管资源与部署计划 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 构建、受管资源与部署计划 Profile 1.0

标识：`uicl.deployment@1.0.0`。状态：规范整合稿。依赖：core, resource, testing。

不可变产物、受管状态、计划、外部授权及执行回执。


## `PlanState`

持久受管状态。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `binding` | `text` | 是 | 否 | 状态存储绑定名，不含秘密。 |
| `locking` | `text` | 是 | 否 | 并发应用锁。 |
| `encryption` | `text` | 否 | 否 | 状态加密要求。 |

## `build`

由源契约生成不可变产物的任务声明。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `source` | `ref` | 是 | 否 | 源契约或源码。 |
| `target` | `ref` | 是 | 否 | 目标工具链。 |
| `artifact_type` | `text` | 是 | 否 | 产物类型。 |
| `embed` | `list<ref>` | 否 | 否 | 明确嵌入的其他构建输出。 |
| `toolchain_binding` | `text` | 是 | 否 | 执行时解析的工具链锁。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |
| `entrypoint` | `text` | 否 | 否 | 输出包内应用入口，由目标适配器验证。 |
| `static_mount` | `text` | 否 | 否 | 嵌入静态产物的应用 URL 前缀，由锁定适配器实现。 |

输出端口：`output: Artifact`，可用阶段 `execution`。

## `managed.resource`

通用受管资源抽象的标准入口；提供方扩展需给 schema。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `provider` | `text` | 是 | 否 | 适配器。 |
| `external_id` | `text` | 否 | 否 | 接管已有资源时的真实 ID。 |
| `desired` | `record` | 是 | 否 | 提供方验证的期望状态。 |
| `managed_fields` | `list<text>` | 是 | 否 | 明确管理的字段。 |
| `on_remove` | `text` | 是 | 否 | 删除策略。 |
| `import_policy` | `text` | 是 | 否 | 已有资源接管规则。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

## `deployment`

将期望状态解析成计划，不自动 apply。

主文本：`name`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 环境名。 |
| `resources` | `list<ref>` | 是 | 否 | 受管资源集合。 |
| `state` | `record<PlanState>` | 是 | 否 | 状态与锁。 |
| `existing_resources` | `text` | 是 | 否 | 碰到已有对象的规则。 |
| `approval` | `text` | 是 | 否 | 计划应用需外部批准。 |
| `gates` | `list<ref>` | 否 | 否 | 必需通过的测试/检查。 |

输出端口：`plan: Plan`，可用阶段 `execution`。

输出端口：`record: ExecutionRecord`，可用阶段 `execution`。

## `plan`

可审阅的执行计划元记录。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `subject` | `ref` | 是 | 否 | 部署/迁移/生成任务。 |
| `expected_state` | `text` | 是 | 否 | 被观察状态版本摘要。 |
| `steps` | `list<record>` | 是 | 否 | 有依赖、前置/后置与恢复分类的步骤。 |
| `approval_binding` | `text` | 是 | 否 | 外部批准服务引用，不是批准本身。 |

## `execution.record`

持久执行回执。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `plan_digest` | `text` | 是 | 否 | 实际计划摘要。 |
| `state` | `text` | 是 | 否 | 执行状态。 |
| `steps` | `list<record>` | 是 | 否 | 逐步骤回执与观察到的外部 ID。 |
| `secrets` | `text` | 是 | 否 | 只允许 redact。 |

## 规则 DEPLOY_PLAN

observe/plan/apply 分开。仅管理显式接管的资源与字段，不因源码缺失而删除外部对象。应用计划时核对状态版本、锁和外部审批，漂移必须重新计划。


验证类别：`planner-runtime`。

## 规则 DEPLOY_RECOVERY

部分成功保留外部身份和回执。网络中断/超时可能 INDETERMINATE，先 observe/reconcile，不盲目重复副作用。应用、结构、数据恢复分别声明。


验证类别：`runtime`。

## 规则 DEPLOY_SINGLE_OWNER

同一资源或字段只有一个明确管理者；生成配置是派生文件，不与控制台/其他工具独立维护同一字段事实来源。批准与签名不是源码字段自我授予。


验证类别：`planner-runtime`。
