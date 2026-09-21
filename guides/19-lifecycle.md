<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_lifecycle "需求、项目与观测 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 需求、项目与观测 Profile 1.0

标识：`uicl.lifecycle@1.0.0`。状态：规范整合稿。依赖：core, application, testing, deployment。

把完整开发交付过程关联起来，不把计划当执行事实。


## `requirement`

可引用的业务需求。

主文本：`title`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 需求摘要。 |
| `statement` | `text` | 是 | 否 | 完整要求。 |
| `acceptance` | `list<text>` | 是 | 否 | 验收标准，不等于已通过。 |

## `decision`

未决需求阻断发布。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `question` | `text` | 是 | 否 | 需要决定的问题。 |
| `options` | `list<text>` | 是 | 否 | 候选。 |
| `resolved` | `bool` | 是 | 否 | 是否已明确解决。 |
| `answer` | `text` | 否 | 否 | 已确认答案。 |

## `project`

跨模块项目入口。

主文本：`title`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 项目名。 |
| `requirements` | `list<ref>` | 是 | 否 | 需求列表。 |
| `design` | `ref` | 是 | 否 | 设计系统。 |
| `app` | `ref` | 是 | 否 | UI 应用。 |
| `service` | `ref` | 是 | 否 | 后端服务。 |
| `database` | `ref` | 是 | 否 | 数据库。 |
| `tests` | `list<ref>` | 是 | 否 | 验收。 |
| `builds` | `list<ref>` | 是 | 否 | 构建声明。 |
| `deployment` | `ref` | 是 | 否 | 部署环境。 |

## `observe`

运行观测契约，不默认收集正文。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `ref` | 是 | 否 | 被观测对象。 |
| `signals` | `list<text>` | 是 | 否 | 日志/指标/追踪。 |
| `payloads` | `text` | 是 | 否 | 禁止敏感正文。 |
| `retention_days` | `int` | 是 | 否 | 保留期限。 |
| `binding` | `text` | 是 | 否 | 授权接收端。 |

## 规则 LIFECYCLE_GATES

未解决 decision、未知检查、缺失工具/模型/身份绑定或未授权的数据出站均阻断发布。不因语法可解析就声称应用可运行。


验证类别：`planner-runtime`。

## 规则 LIFECYCLE_OBSERVE

日志/追踪默认不记录请求正文、凭据、模型完整上下文；没有授权接收端不上传。验证/执行状态必须诚实区分声明、计划、实际回执。


验证类别：`runtime`。
