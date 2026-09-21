<!-- uicl:markdown 1.0 -->

<!-- uicl
document #quickstart "UICL 1.0 入门"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# UICL 1.0 入门

## 1. 最小文档

```uicl
uicl "1.0"

app "你好"
  ui.text "我的第一个 UICL 应用"
```

## 2. 增加状态和事件

```uicl
uicl "1.0"

app #counter "计数器"
  state #count
    value: 0
  ui.text
    value: $count
  ui.button #increase "加一"
    on click
      set
        target: @count
        value: = $count + 1
```

## 3. 数组和对象使用清单写法

```uicl
uicl "1.0"

app #notes_app "随手记"
  collection #notes
    storage: device
    operations:
      - read
      - create
    field "content"
      type: text
      min: 1
  form #editor "新增便签"
    create: @notes
  list #entries
    source: @notes
    empty_text: "尚无便签。"
```

## 4. 文档与界面不是同一种对象

普通文章使用 Markdown-hosted UICL；普通网页使用 HTML-hosted UICL。image 是生成规格，ui.image 是显示已有图像。不能把相同词汇跨上下文随意解释。

## 5. 下一步阅读

设计与 UI 见 examples/04-07；原生与部署见 08-11；AIGC 见 12-18；Agent 与混合程序见 19-22；文件与数据集见 23-25；完整开发流程见 examples/fullstack/project.uicl。

本包提供的是规范和检查原型。没有已实现的 uicl dev/build/deploy 或 Semaquil 产品命令。真实任务需要对应运行时和宿主绑定。
