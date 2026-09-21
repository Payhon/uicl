<!-- uicl:markdown 1.0 -->

<!-- uicl
document #domain_semantics "UICL 领域执行语义补充"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# UICL 领域执行语义补充

## 如何使用这份补充

这些规则补齐前面对跨领域接口的推演，来源类别为“1.0 编辑细化”。它们是待适配器实现和验证的规范，不是本包已运行结果。

### 输入、参数与默认值

parameter 由宿主提供并经声明 type 验证；secret 不导出 $ 文字绑定。schema.default 的注入属于展开阶段，语法/形状检查保留“缺失”与“已给默认值”的区别。生产锁固定全部已使用目录和默认 recipe。

### query、resource、list

resource 的 $id 快照包含 items、loading、error、next_cursor；错误字段是脱敏错误或 null。list 的子节点模板逐项获得 $item，只读且 key 默认 id。每次列表刷新带版本号；新请求获胜。for/each 的 as 不能遮蔽保留上下文。

### action 与生成结果

create/update/call/invoke 的 #id 在顺序上下文中导出 $id 结果；不同节点的端口 result/output 以各自 schema 为准。form.create=action 必须与该 action 的命名输入生成字段一一匹配，不能提交 owner 等额外字段。未返回型 action 只能在输出 unit 的正常终点结束。

### HTTP 与类型化业务映射

method/path 必须唯一配对；action 与 query 互斥。GET/HEAD 不以隐含副作用写模型。auth/public 不覆盖内部 allow。rate_limit 为 HTTP 适配器封闭输入：requests、window_seconds 为正整数，key 为 ip 或 actor，actor 要求已认证。结构检查的 generic record 不表示跳过此业务检查。

### PostgreSQL 迁移步骤最小注册表

steps.operation 可为 create_table、add_column、rename_column、create_index、drop_column。每个步骤必须指定 target；add/rename/drop 需明确字段名和必要类型，create_index 需 name/build 及已在表契约中定义的索引。删除受 destructive 门禁控制；不支持的操作报错，不交给模型生成任意 SQL。数据库范围扩大时新增版本化操作。

### 图像与动态 UI 约束

size 必须两个正整数；region 是左上原点归一化 [x,y,w,h]，x/y 非负、w/h 为正且矩形不得越出 1。design.layer.bounds 使用画布逻辑坐标。generator output=Artifact<ui.tree> 时 constraints 必須声明 components/max_nodes/max_depth/actions 并完成类型检查；image 输出时指定 size/media。其他字段需要相应输出 Profile 注册，不能隐式接受任意约束。

### 原生 API 与移动权限

原生 interface 清单必须给句柄所有权、回调生命周期、取消和错误映射。Windows 示例仅是 capability contract；附带的 C# 源码是明确未实现的接口夹具。Mobile 接口必须把 CANCELLED、PERMISSION_DENIED、UNAVAILABLE 等固定结果映射给应用，不将失败伪装成空图片。

### 全流程示例的构建边界

go_api_with_static_bundle 是示例要求的宿主构建适配器，不是已发布实现。其契约应生成包含 bin/api 的压缩包，将嵌入的 Web 产物在 / 下服务，并使 API /api/* 与 /health 不被 SPA 回退覆盖。实际绑定缺失即停止。

PostgreSQL schema 管理凭据与应用 DATABASE_URL 必须指向同一部署环境/数据库，但应用角色应更小权限；示例声明两个句柄并不证明这一条件。SSH 服务账户、发布目录和 sudo 规则要求前置验证。示例只绑定 127.0.0.1:8080；公网 TLS、域名、网关和云安全组需另外显式声明，不默认暴露服务。

### 部署资源先后与云资源

值/资源句柄依赖可形成计划边；depends_on 提供额外显式前置。Cloudflare Worker 的 DB/FILES 绑定要求相应资源存在。ECS release 依赖目标主机与显式数据库准备；资源列表顺序仅保留作者表达顺序，不授予依次执行语义。

### 证据与规范未覆盖项

没有提供方/SQL/设备运行器时返回 NOT_EVALUATED 或 UNSUPPORTED_CAPABILITY；不使用通用自然语言执行器吞掉未知字段。对新领域而言，需要扩展 Profile、规则和适配器，而不是在 Core 中增加一组私有标点。
