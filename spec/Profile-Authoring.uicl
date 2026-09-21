<!-- uicl:markdown 1.0 -->

<!-- uicl
document #profile_authoring "UICL 领域 Profile 作者标准"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# UICL 领域 Profile 作者标准

## 1. Profile 与实现的区别

Profile 规定一个领域的节点、输入输出、约束与错误。Adapter 实现这些规则。只有 schema 不能宣称功能可执行；没有节点形状的任意 prompt 也不构成稳定领域规范。

第三方命名空间不得占用标准含义，不改变 #、@、$、=、缩进和集合语义。Profile 的依赖必须锁定并可离线检查；执行依赖不允许自行下载后运行未知代码。

## 2. 一个真实的扩展定义

下面是包内的完整自定义 Profile，使用 Meta Profile 描述自己：

```uicl
uicl "1.0"

profile #measurement_profile "测量数据领域"
  id: "example.measurement"
  version: "1.0.0"
  core: "1.0"
  status: "example-extension"
  summary: |
    演示第三方 Profile 的标准写法；只描述测量记录，不读取真实设备。
  requires:
    - "uicl.core@1.0.0"
  origin: "1.0-profile-authoring-example"
  implementation: "shape-only-example"
  namespaces:
    - "lab"
  schema #sample_schema "lab.sample"
    summary: "带单位的测量记录。"
    primary: "label"
    identity: "required"
    children: []
    unknown_properties: "reject"
    effects: []
    category: "data"
    property "label"
      type: "text"
      description: "测量名称。"
      required: true
      expression: false
    property "value"
      type: "decimal"
      description: "精确十进制测量值。"
      required: true
      expression: false
    property "unit"
      type: "text"
      description: "单位。"
      required: true
      expression: false
      enum:
        - "V"
        - "A"
        - "W"
    port "value"
      type: "decimal"
      phase: "compile"
  rule #measurement_rule "MEASUREMENT_UNIT"
    level: "MUST"
    statement: |
      单位是测量语义的一部分；读取此记录不读取设备。需要转换时必须使用明确且版本化的转换器。
    verification: "domain"
    origin: "1.0-profile-authoring-example"
```

## 3. 它的使用文档

```uicl
uicl "1.0"

lab.sample #voltage "电压"
  value: 12.6
  unit: "V"
```

编辑器仅加载明确指定的扩展；示例中的 lab.sample 没有实际设备调用。

```bash
python tools/check.py \
  examples/profile-authoring/measurement-profile.uicl \
  examples/profile-authoring/measurement-example.uicl \
  --extra-profile examples/profile-authoring/measurement-profile.uicl
```

## 4. 每个新领域必须回答的问题

节点是数据、规格、产物、状态、能力还是受管资源？哪些字段必需，主文本映射到哪里？引用类型与输出端口是什么？参数含表达式时在哪个环境求值？副作用与秘密如何约束？失败能否重试，结果不确定如何处理？哪些规则可机器验证，哪些需要人工证据？目标不支持时如何失败？

不应把所有字段都写为无约束 record/value。确有开放参数区时，必须指向实例化后的业务或适配器输入契约，缺失则阻断执行。本包记录的通用形状检查不会假装完成这一步。

## 5. 版本与兼容

添加无默认语义影响的可选元数据通常可作为兼容修订；改变 primary、默认值、权限、错误重试、副作用或序列顺序属于行为变更，必须显式版本化。标准节点不得用新 Profile 同名覆盖。

每个 Profile 应提供正例、非法字段、缺失依赖、权限拒绝、错误恢复和目标不支持的向量。声明源文和真实运行报告分开。自描述规则不接受其输入文档自动扩大宿主权限。

## 6. 第三方生态与平台组合

AKML/SVML 可以定义自己的组合目录、默认组件、适配器与平台词汇，但保持 UICL Core 不变。优先提供通用 .uicl 导出，并标明平台专属、不可移植或尚未实现的能力。不强制所有 UICL 实现采用 Semaquil。
