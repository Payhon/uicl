<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_testing "测试与验收证据 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 测试与验收证据 Profile 1.0

标识：`uicl.testing@1.0.0`。状态：规范整合稿。依赖：ui, backend, resource。

测试声明、夹具、断言和验证记录；不会在读取时执行。


## `test`

有隔离级别与目标的测试。

主文本：`title`；身份：`required`；允许子节点：given, when, then, expect。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `title` | `text` | 是 | 否 | 测试意图。 |
| `kind` | `text` | 是 | 否 | 测试种类。 |
| `verifies` | `list<ref>` | 否 | 否 | 关联的需求或规则。 |
| `isolation` | `text` | 是 | 否 | 测试实例隔离。 |
| `timeout_ms` | `int` | 否 | 否 | 测试超时。 |

输出端口：`evidence: Evidence`，可用阶段 `execution`。

## `given`

测试阶段，动作仅作用于隔离环境。

主文本：`None`；身份：`optional`；允许子节点：test.reset, test.open, test.fill, test.submit, test.click, test.execute, test.fixture, test.actor。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `when`

测试阶段，动作仅作用于隔离环境。

主文本：`None`；身份：`optional`；允许子节点：test.reset, test.open, test.fill, test.submit, test.click, test.execute, test.fixture, test.actor。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `test.reset`

清空隔离测试数据，禁止生产实例。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `resource` | `ref` | 是 | 否 | 测试数据源。 |

## `test.open`

打开指定页面。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `page` | `ref` | 是 | 否 | 目标页面。 |

## `test.fill`

向指定表单字段输入。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `ref` | 是 | 否 | input/form。 |
| `field` | `text` | 否 | 否 | form 字段名。 |
| `value` | `value` | 是 | 否 | 输入值。 |

## `test.submit`

提交指定表单。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `ref` | 是 | 否 | 表单。 |

## `test.click`

点击交互对象。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `ref` | 是 | 否 | 按钮。 |

## `test.execute`

调用动作并捕获结果或错误。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `action` | `ref` | 是 | 否 | 后端动作。 |
| `args` | `record` | 是 | 否 | 输入。 |

输出端口：`result: TestResult`，可用阶段 `execution`。

## `test.fixture`

隔离数据夹具；不绕过被测业务授权路径。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `model` | `ref` | 是 | 否 | 模型。 |
| `rows` | `list<record>` | 是 | 否 | 初始记录。 |

## `test.actor`

注入测试身份，生产禁用。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `id` | `text` | 是 | 否 | 测试用户 ID。 |
| `roles` | `list<text>` | 否 | 否 | 角色集合。 |

## `expect`

测试断言；未评估不是通过。

主文本：`None`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `target` | `ref` | 否 | 否 | 目标对象或测试结果。 |
| `contains` | `text` | 否 | 否 | 包含的可见文本。 |
| `count` | `int` | 否 | 否 | 数量。 |
| `equals` | `value` | 否 | 否 | 结构/标量期望值。 |
| `error` | `text` | 否 | 否 | 预期错误码。 |
| `test` | `bool` | 否 | 允许，仍需业务检查 | 布尔表达式。 |

## `check`

登记验证器的硬检查。

主文本：`rule`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `rule` | `text` | 是 | 否 | 验证器标识。 |
| `target` | `ref` | 否 | 否 | 待验证对象；缺省为外层产物。 |
| `equals` | `value` | 否 | 否 | 期望值。 |
| `min` | `number` | 否 | 否 | 下界。 |
| `max` | `number` | 否 | 否 | 上界。 |
| `required` | `bool` | 否 | 否 | 是否阻断发布。 |

输出端口：`evidence: Evidence`，可用阶段 `execution`。

## `require`

所有必需检查必须 PASS 才满足。

主文本：`None`；身份：`optional`；允许子节点：check。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|

## `review`

自然语言评审，必须记录评估者及证据。

主文本：`None`；身份：`optional`；允许子节点：criterion。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `evaluator` | `text` | 否 | 否 | 评估器绑定或 human。 |

## `criterion`

自然语言评价维度。

主文本：`text`；身份：`optional`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `text` | `text` | 是 | 否 | 评审标准，不会直接编译为确定性检查。 |

## `evidence`

绑定真实检查输入和结果的记录。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `subject_digest` | `text` | 是 | 否 | 被测产物摘要。 |
| `contract_digest` | `text` | 是 | 否 | 规格摘要。 |
| `validator` | `text` | 是 | 否 | 验证器与版本。 |
| `status` | `text` | 是 | 否 | 实际状态。 |
| `report` | `ref` | 否 | 否 | 具体报告。 |

## 规则 TEST_ISOLATION

reset/fixture/actor 注入只能作用于隔离测试实例。测试文件不能授予生产权限；自动生成的测试不代替独立业务验收。


验证类别：`runtime`。

## 规则 TEST_EVIDENCE

证据绑定实际产物、契约、验证器版本及输入摘要。旧产物的通过记录不能用于新产物。NOT_EVALUATED/ERROR 不满足 require；被测 Agent 不能自行删除门禁。


验证类别：`runtime`。
