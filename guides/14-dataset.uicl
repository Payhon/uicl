<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_dataset "训练样本与来源血缘 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 训练样本与来源血缘 Profile 1.0

标识：`uicl.dataset@1.0.0`。状态：规范整合稿。依赖：resource, testing。

记录规格、产物、来源、许可、评估及拆分分组。


## `dataset`

样本集合与拆分原则。

主文本：`name`；身份：`required`；允许子节点：sample。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 否 | 否 | 数据集名。 |
| `purpose` | `text` | 是 | 否 | 明确用途。 |
| `split_policy` | `text` | 是 | 否 | 以共同来源分组防泄漏。 |
| `permission` | `text` | 是 | 否 | 许可核验状态。 |
| `loader` | `text` | 是 | 否 | 还原原生模态的加载方式。 |

## `sample`

有可追踪输入输出的一条样本。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `source_group` | `text` | 是 | 否 | 同源文档/图像/投影共享的分组。 |
| `input` | `ref` | 是 | 否 | 任务规格或原文资源。 |
| `output` | `ref` | 是 | 否 | 实际产物，不能引用未执行生成任务假装存在。 |
| `license` | `text` | 是 | 否 | 内容权利/许可，unknown 不允许训练。 |
| `consent` | `text` | 是 | 否 | 必要同意证据或核验状态。 |
| `sensitivity` | `text` | 是 | 否 | 敏感性。 |
| `evidence` | `list<ref>` | 否 | 否 | 验证记录。 |

## 规则 DATASET_RIGHTS

采用 .uicl 或 HEXL 不改变用户内容权利，不表示允许训练、公开或转让。许可、同意未知时必须阻断数据发布/训练。原始来源同组拆分。


验证类别：`governance-runtime`。

## 规则 DATASET_MODALITY

交换层可以文本封装；训练层根据模态解码，不将 Base64 当作视觉语义。统一格式提升模型能力属于待实验假设。


验证类别：`loader-evaluation`。
