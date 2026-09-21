<!-- uicl:markdown 1.0 -->

<!-- uicl
document #profile_index "UICL 1.0 Profile 索引"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# UICL 1.0 Profile 索引

每个 Profile 都有机器可读的 `.uicl` 定义和从其生成的文档。字段、主文本、嵌套记录、子节点和规则以 `profiles/` 为源，不维护两份互相独立的定义。

| Profile | 领域 | 节点 / 记录 | 规范源 |
|---|---|---:|---|
| `uicl.meta` | Profile 与语言自描述 | 16 / 0 | [00-meta.uicl](../profiles/00-meta.uicl) |
| `uicl.core` | 核心类型、模块与执行上下文 | 11 / 1 | [01-core.uicl](../profiles/01-core.uicl) |
| `uicl.resource` | 文件、产物与任意字节 | 4 / 0 | [02-resource.uicl](../profiles/02-resource.uicl) |
| `uicl.design` | 设计系统与设计稿 | 6 / 1 | [03-design.uicl](../profiles/03-design.uicl) |
| `uicl.ui` | 界面与交互 | 23 / 0 | [04-ui.uicl](../profiles/04-ui.uicl) |
| `uicl.backend` | 后端业务与 API | 25 / 0 | [05-backend.uicl](../profiles/05-backend.uicl) |
| `uicl.application` | 应用组合与渐进默认 | 3 / 0 | [06-application.uicl](../profiles/06-application.uicl) |
| `uicl.postgresql` | PostgreSQL 关系结构与迁移 | 6 / 5 | [07-postgresql.uicl](../profiles/07-postgresql.uicl) |
| `uicl.testing` | 测试与验收证据 | 17 / 0 | [08-testing.uicl](../profiles/08-testing.uicl) |
| `uicl.deployment` | 构建、受管资源与部署计划 | 5 / 1 | [09-deployment.uicl](../profiles/09-deployment.uicl) |
| `uicl.aigc` | 生成内容规格与任务 | 18 / 1 | [10-aigc.uicl](../profiles/10-aigc.uicl) |
| `uicl.agent` | Agent 定义与工具契约 | 4 / 2 | [11-agent.uicl](../profiles/11-agent.uicl) |
| `uicl.hybrid` | 推断、生成槽位与沙箱子应用 | 5 / 0 | [12-hybrid.uicl](../profiles/12-hybrid.uicl) |
| `uicl.document` | 原生文档与结构化文档规格 | 3 / 0 | [13-document.uicl](../profiles/13-document.uicl) |
| `uicl.dataset` | 训练样本与来源血缘 | 2 / 0 | [14-dataset.uicl](../profiles/14-dataset.uicl) |
| `uicl.native` | 桌面原生 API 与系统操作 | 1 / 2 | [15-native.uicl](../profiles/15-native.uicl) |
| `uicl.mobile` | 移动端目标和设备能力 | 2 / 1 | [16-mobile.uicl](../profiles/16-mobile.uicl) |
| `uicl.cloudflare` | Cloudflare 受管基础设施 | 5 / 0 | [17-cloudflare.uicl](../profiles/17-cloudflare.uicl) |
| `uicl.host` | SSH 主机操作与 ECS 发布 | 2 / 6 | [18-host.uicl](../profiles/18-host.uicl) |
| `uicl.lifecycle` | 需求、项目与观测 | 4 / 0 | [19-lifecycle.uicl](../profiles/19-lifecycle.uicl) |
