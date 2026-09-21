# AppMark 语言规范 v1.0

> 全栈应用描述语言。人写意图，编译器补实现，AI 做安全修改。

---

## 目录

1. 概述
2. 设计目标
3. 文件与词法
4. 程序结构
5. 类型系统
6. 设计层 design
7. 数据层 data
8. 接口层 api
9. 前端层 front
10. 测试层 test
11. 智能体层 agent
12. 部署层 deploy
13. 表达式与语句
14. 组件与组合
15. 语义图与工具链
16. 错误模型
17. 编译目标
18. 兼容与逃逸

---

## 1. 概述

**AppMark** 是一门用于描述全栈应用的语言。文件后缀 `.appm`。

一份 AppMark 文档同时描述：

- 设计系统（design）
- 数据模型（data）
- 接口契约（api）
- 前端界面（front）
- 测试用例（test）
- 智能体意图（agent）
- 部署拓扑（deploy）

编译器将这些声明投影为 SQL、TypeScript、SwiftUI、Flutter、OpenAPI、容器配置、测试代码等。

AppMark 不是通用编程语言。复杂算法通过类型化 `fn` 逃逸块完成，其余环节统一在同一张语义图中。

---

## 2. 设计目标

1. **人类可读**：缩进即结构，无闭合标签，无冗余括号。
2. **人类易写**：默认值丰富，常见操作一行完成。
3. **AI 可理解**：稳定 ID、显式契约、结构化意图、可验证修改。
4. **单一事实源**：设计、前端、后端、存储是同一图的四个投影。
5. **渐进增强**：新手只写页面，进阶加状态，专家写策略与迁移。
6. **跨端编译**：一套源码生成 Web、iOS、Android、边缘函数。
7. **可 diff、可影响分析、可回滚**。

---

## 3. 文件与词法

### 3.1 编码

- UTF-8。
- 换行 `\n` 或 `\r\n`，编译器统一为 `\n`。
- 缩进用 2 个空格。Tab 自动转换为 2 空格。

### 3.2 注释

```appmark
// 单行注释
/* 块注释 */
```

结构化注释使用字段形式，见 §11：

```appmark
intent "极简商店"
why "移动端优先"
```

### 3.3 标识符

- 正则：`[A-Za-z_][A-Za-z0-9_]*`
- 组件名首字母大写：`Card`、`TodoList`
- 变量、动作、字段小写驼峰：`draft`、`addTodo`
- 稳定 ID：`@id` 后接点分命名空间：`@id data.product`

### 3.4 字面量

```appmark
42           // 整数
3.14         // 小数
true false   // 布尔
"hello"      // 字符串
'hello'      // 字符串（单引号）
`模板 {name}` // 模板字符串
#2563eb      // 颜色
null
```

### 3.5 运算符

```text
+ - * / %           算术
== != < <= > >=     比较
&& || !             逻辑
??                  空值合并
? :                 三元
=                   赋值
++ --               自增自减
```

---

## 4. 程序结构

顶层块按任意顺序出现。推荐顺序：

```appmark
app 名称
  intent "..."
  constraint "..."

design
data
api
front
test
agent
deploy
```

每个块使用 2 空格缩进，子块继续缩进。

### 4.1 块头

```appmark
app Shop @id shop
  intent "极简商店，浏览商品、下单、管理库存"
```

`app` 是根，`@id` 是稳定标识。

---

## 5. 类型系统

### 5.1 基础类型

| 类型 | 说明 | 示例 |
|---|---|---|
| `id` | 主键 | `id: id` |
| `text` | 字符串 | `name: text` |
| `int` | 整数 | `stock: int` |
| `float` | 小数 | `price: float` |
| `money` | 货币，分为单位存储 | `price: money` |
| `bool` | 布尔 | `active: bool` |
| `date` | 日期 | `created: date` |
| `time` | 时间 | `at: time` |
| `datetime` | 日期时间 | `at: datetime` |
| `json` | 任意 JSON | `meta: json` |
| `enum(...)` | 枚举 | `status: enum(draft, active)` |
| `ref X` | 外键引用 | `user: ref User` |
| `list<T>` | 列表 | `items: list<OrderItem>` |
| `map<K,V>` | 映射 | `meta: map<text, text>` |

### 5.2 修饰符

```appmark
name: text required unique
email: text required unique index
stock: int = 0
total: money computed sum(items.price * items.qty)
```

- `required`：非空
- `unique`：唯一
- `index`：建索引
- `= 默认值`
- `computed 表达式`：只读派生

---

## 6. 设计层 design

### 6.1 设计令牌

```appmark
design
  tokens
    color
      primary = #2563eb
      bg = #f8fafc
      text = #0f172a
      muted = #64748b
    radius
      sm = 6
      md = 12
      lg = 20
    spacing
      sm = 8
      md = 16
      lg = 24
    font
      body = "Inter"
      mono = "JetBrains Mono"
      scale = [12, 14, 16, 20, 28, 40]
```

### 6.2 主题

```appmark
  theme light
    bg = $color.bg
    text = $color.text

  theme dark
    bg = #0b1220
    text = #e2e8f0
```

主题变量通过 `$` 前缀引用：`$primary`、`$color.primary`。

### 6.3 样式

```appmark
  style card
    padding = $spacing.md
    radius = $radius.md
    bg = $surface
    shadow = sm
```

使用：`view style=card`。

### 6.4 可访问性

```appmark
  a11y
    minContrast = 4.5
    focusRing = $primary
    reduceMotion = true
```

---

## 7. 数据层 data

### 7.1 实体

```appmark
data
  entity Product @id data.product
    id: id
    name: text required
    price: money required
    stock: int = 0
    status: enum(draft, active) = draft
    createdAt: datetime = now()
    index status
    policy
      read: true
      write: auth.role == admin
```

### 7.2 关系

```appmark
  entity Order @id data.order
    id: id
    user: ref User required
    items: list<OrderItem>
    total: money computed sum(items.price * items.qty)
    status: enum(pending, paid, shipped) = pending
    policy
      read: auth.user == user || auth.role == admin
      create: auth.required
      update: auth.user == user && status == pending
```

关系类型：

- `ref X`：多对一
- `list<X>`：一对多
- `ref X many`：多对多
- `ref X one`：一对一

### 7.3 策略

`policy` 下每个动作是一个表达式：

```appmark
    policy
      read: true
      create: auth.required
      update: auth.user == user
      delete: auth.role == admin
```

策略编译为行级权限（RLS）、接口鉴权、前端可见性判断。

### 7.4 迁移

```appmark
  migrate
    from 1 to 2
      add Product.rating: float = 0
      rename Product.desc -> description
```

迁移影响通过 `appm impact` 分析。

---

## 8. 接口层 api

### 8.1 查询

```appmark
api
  query activeProducts @id api.activeProducts
    from Product where status == active
    sort createdAt desc
    return Product[]
```

### 8.2 变更

```appmark
  mutation placeOrder(items) @id api.placeOrder
    auth required
    validate items.length > 0
    stock.checkAndDecrease(items)
    insert Order { user: auth.user, items }
    return Order
```

### 8.3 订阅

```appmark
  subscribe orderUpdates @id api.orderUpdates
    from Order where user == auth.user
    return Order
```

### 8.4 外部接口

```appmark
  external payment @id api.payment
    method POST
    url "https://api.stripe.com/v1/charges"
    auth bearer env.STRIPE_KEY
    input { amount: money, currency: text }
    output { id: text, status: text }
```

### 8.5 错误码

```appmark
  errors
    OUT_OF_STOCK "库存不足"
    PAYMENT_FAILED "支付失败"
```

在动作中：

```appmark
    if product.stock < qty
      throw OUT_OF_STOCK
```

---

## 9. 前端层 front

### 9.1 页面

```appmark
front
  page Shop @id page.shop
    title "商店"
    data products = query api.activeProducts
    state cart = []

    view padding=$spacing.md gap=$spacing.md
      text "商店" size=28 bold
      list items={products} key={id}
        item
          Card
            row align=center gap=8
              text "{name}" flex=1
              text "¥{price}" bold
              button "加入购物车" onTap={ cart.push(product) }
      if cart.length > 0
        button "结算" onTap={ placeOrder(cart) }
```

### 9.2 内置组件

布局：`view`、`row`、`column`、`stack`、`grid`、`list`、`item`、`scroll`、`spacer`、`divider`  
内容：`text`、`image`、`icon`、`video`  
交互：`button`、`input`、`checkbox`、`radio`、`select`、`slider`、`switch`、`form`  
反馈：`spinner`、`toast`、`modal`、`sheet`、`tooltip`

### 9.3 通用属性

| 属性 | 类型 | 说明 |
|---|---|---|
| `padding` `margin` | 数字 / 令牌 | 间距 |
| `gap` | 数字 | 子元素间距 |
| `bg` | 颜色 | 背景 |
| `color` | 颜色 | 前景 |
| `radius` | 数字 | 圆角 |
| `shadow` | `sm` `md` `lg` | 阴影 |
| `size` | 数字 | 字号 |
| `bold` `italic` `strike` | 布尔简写 | 字重样式 |
| `flex` | 数字 | 弹性 |
| `align` | `start` `center` `end` | 对齐 |
| `width` `height` | 数字 / `fill` | 尺寸 |
| `maxWidth` `minWidth` | 数字 | 约束 |
| `if` | 表达式 | 条件渲染 |
| `for` | `item in list` | 循环 |

### 9.4 数据绑定

```appmark
input bind={draft}          // 双向
text value={name}            // 单向
checkbox bind={done}
select bind={status} options={[...]}
```

### 9.5 事件

```appmark
button "保存" onTap={ save() }
input onChange={ e => name = e.value }
form onSubmit={ submit() }
view onLongPress={ showMenu() }
```

### 9.6 状态与计算

```appmark
  state todos = []
  state draft = ""
  computed left = todos.filter(t => !t.done).length
```

### 9.7 路由

```appmark
  route "/" -> Shop
  route "/item/:id" -> Detail
  route "/admin" -> Admin guard={ auth.role == admin }
```

页面中：

```appmark
    button "详情" onTap={ navigate("/item/" + id) }
    button "返回" onTap={ back() }
```

### 9.8 生命周期

```appmark
  page Detail
    onLoad
      item = await GET api.item(route.id)
    onUnload
      cache.clear()
```

---

## 10. 测试层 test

```appmark
test
  test "库存不足不能下单" @id test.outOfStock
    given product stock = 0
    when placeOrder([product])
    expect error OUT_OF_STOCK

  test "下单后库存减少" @id test.stockDecrease
    given product stock = 10
    when placeOrder([product, qty=2])
    expect product.stock == 8

  test "页面显示购物车数量" @id test.cartBadge
    given cart = [p1, p2]
    when render page.shop
    expect text "2" exists
```

### 10.1 断言

```appmark
expect <表达式>
expect error CODE
expect text "..." exists
expect navigate to "/..."
expect api.placeOrder called 1 times
```

### 10.2 测试类型

- `unit`：纯函数、动作
- `integration`：api + data
- `e2e`：front + api + data
- `visual`：截图对比
- `a11y`：可访问性

默认根据 `@id` 命名空间自动分类。

---

## 11. 智能体层 agent

`agent` 块是机器可读的意图、约束、示例和任务边界。

```appmark
agent
  intent "移动端优先；价格用分存储；库存扣减必须原子"

  constraint "前端不能直接改库存"
  constraint "所有金额字段使用 money 类型"
  constraint "任何写操作必须有 policy"

  example
    input "给订单加退款"
    expect
      add data.Refund
      add api.refundOrder
      add page.Refund
      add test.refundFlow

  scope
    allow front/*
    allow api/*
    deny data.product.price
```

### 11.1 字段

| 字段 | 说明 |
|---|---|
| `intent` | 自然语言意图，人类和 AI 都读 |
| `constraint` | 硬约束，违反则编译失败 |
| `why` | 设计理由，影响分析时展示 |
| `example` | 输入输出示例，用于 AI 生成和验证 |
| `scope` | Agent 可修改范围 |
| `review` | 需要人工审核的变更类型 |

---

## 12. 部署层 deploy

```appmark
deploy
  env dev
    db = local
    scale = 1

  env prod
    db = postgres
    region = "us-east-1"
    scale = auto(1..10)
    secrets
      STRIPE_KEY = vault("stripe/prod")
```

编译为容器配置、云函数、CDN、数据库迁移计划。

---

## 13. 表达式与语句

### 13.1 表达式

安全 JS 子集，无 `eval`、无原型污染、无全局副作用。

```appmark
count + 1
items.filter(t => !t.done).length
name.trim()
auth.user == user
status == active ? "上架" : "草稿"
value ?? "默认"
```

### 13.2 语句

```appmark
action addTodo
  if draft.trim()
    todos.push({ id: now(), title: draft.trim(), done: false })
    draft = ""

action save
  try
    await POST "/api/user" { name: name }
    toast "保存成功"
  catch e
    toast e.message
  finally
    loading = false
```

支持：`if / else`、`for`、`while`、`try / catch / finally`、`await`、`return`、`throw`。

---

## 14. 组件与组合

### 14.1 定义

```appmark
component Card(title, subtitle = "")
  view padding=16 radius=12 bg=$surface shadow=sm
    text title bold
    text subtitle color=$muted if={subtitle}
    slot
```

### 14.2 使用

```appmark
Card title="你好"
  text "这是卡片内容"
```

### 14.3 具名插槽

```appmark
component Panel
  slot header
  slot body

Panel
  header
    text "标题"
  body
    text "内容"
```

### 14.4 组件库

```appmark
library ui
  component Button ...
  component Card ...
```

通过 `import "lib/ui.appm"` 引入。

---

## 15. 语义图与工具链

### 15.1 语义图

编译后每个节点有：

```text
id, kind, type, deps, policies, tests, intent, constraints, locations
```

编译器、LSP、AI Agent 都读这张图。

### 15.2 双表示

- **人类视图**：缩进简写。
- **规范视图**：展开所有默认值、策略、类型、依赖。

```bash
appm expand app.appm > app.ir.json
```

### 15.3 命令

```bash
appm fmt            格式化
appm check          类型、策略、约束检查
appm build          编译到目标平台
appm dev            本地热更新
appm test           运行测试
appm impact <id>    影响分析
appm diff           语义 diff
appm migrate        生成迁移
appm deploy         部署
appm agent plan "给订单加退款"   Agent 生成计划
```

### 15.4 影响分析示例

```bash
$ appm impact data.product.stock
→ api.placeOrder (uses)
→ page.shop (displays)
→ test.stockDecrease (asserts)
→ migrate 1→2 (if changed)
```

---

## 16. 错误模型

### 16.1 编译错误

```text
第 12 行：未知组件 buton，是否想用 button？
第 18 行：size 需要数字，但得到字符串 "24"。是否想写 size=24？
第 24 行：写操作必须有 policy。请在 entity Order 下添加 policy.update。
```

### 16.2 运行时错误

```appmark
errors
  OUT_OF_STOCK "库存不足" code=409
  PAYMENT_FAILED "支付失败" code=402
```

`throw OUT_OF_STOCK` 会带上 code、message、trace。

### 16.3 约束违规

```text
约束违规：constraint "前端不能直接改库存"
位置：page.shop line 42
建议：通过 api.placeOrder 修改库存
```

---

## 17. 编译目标

| 目标 | 产物 |
|---|---|
| Web | React / Vue / Svelte + TypeScript |
| iOS | SwiftUI |
| Android | Jetpack Compose |
| 跨端 | Flutter |
| 小程序 | 微信 / 支付宝小程序 |
| 后端 | Node / Go / Python / Rust |
| 数据库 | PostgreSQL / MySQL / SQLite / MongoDB |
| API | REST / GraphQL / tRPC / gRPC |
| 部署 | Docker / K8s / 边缘函数 / Serverless |
| 文档 | Markdown / OpenAPI / 设计系统手册 |

编译目标通过 `appm.config` 配置：

```appmark
config
  targets = [web, ios, postgres, node]
  output = "./dist"
```

---

## 18. 兼容与逃逸

### 18.1 类型化逃逸

复杂算法用 `fn` 块：

```appmark
fn recommend(user, products) -> list<Product>
  lang typescript
  // 这里可以写完整 TS，编译器做类型边界检查
  return products
    .filter(p => p.stock > 0)
    .sort((a, b) => score(b, user) - score(a, user))
    .slice(0, 10)
```

`fn` 是类型安全的逃逸舱，可与 AppMark 表达式互相调用。

### 18.2 嵌入现有代码

```appmark
import "lib/ui.appm"
import fn "src/utils.ts" as utils
import data "schema.prisma" as legacy
```

### 18.3 渐进迁移

- HTML 项目：`appm import html index.html` 生成初版。
- OpenAPI：`appm import openapi api.yaml` 生成 `api` 块。
- Prisma / SQL：`appm import sql schema.sql` 生成 `data` 块。

---

## 附录 A：最小示例

```appmark
app Hello
  intent "最小可运行应用"

front
  page Home
    state name = "世界"
    view padding=24 gap=12
      text "你好，{name}" size=24 bold
      input bind={name} placeholder="输入名字"
      button "打招呼" onTap={ toast "你好，" + name }
```

编译到 Web、iOS、Android，无需其他文件。

---

## 附录 B：关键字表

```text
app design data api front test agent deploy
entity policy index migrate
query mutation subscribe external errors
page route view row column stack grid list item scroll spacer divider
text image icon video
button input checkbox radio select slider switch form
spinner toast modal sheet tooltip
state computed action onLoad onUnload
component slot library import fn config
intent constraint why example scope review
if else for while try catch finally await return throw
```

---

## 附录 C：设计原则速查

1. 缩进即结构，无闭合标签。
2. 默认值丰富，一行可用。
3. 稳定 ID 定位一切。
4. 契约显式：类型、权限、错误、副作用。
5. 意图结构化，AI 可读。
6. 双表示：人类简写，机器展开。
7. 可 diff、可影响分析、可回滚。
8. 渐进增强，复杂逻辑逃逸到 `fn`。
9. 单一事实源，多端投影。
10. 错误提示像人话，并给出修复建议。

---

**AppMark v1.0 — 人写意图，编译器补实现，AI 做安全修改。**