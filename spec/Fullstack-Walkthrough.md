<!-- uicl:markdown 1.0 -->

<!-- uicl
document #fullstack_guide "全流程应用示例说明"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# 全流程应用示例说明

## 目录与引用关系

```text
project.uicl
  requirements.uicl
  design.uicl
  backend.uicl
  database.uicl
  ui.uicl
  tests.uicl
  build.uicl
  deploy.uicl
```

## 整体契约

用户通过 Web 表单创建私有待办。UI 引用后端 create action 和 query；后端模型强制 owner 隔离、标题验证和版本检查；数据库映射包含全部系统字段。测试引用同一 UI 与后端对象，不复制一套独立的业务定义。

构建先生成 Web，再生成含静态资源的 Linux Go 包；部署通过已有 Linux ECS 的 SSH 连接、固定主机指纹和受控 sudo 发布服务，校验 /health 后保留回执。数据库与程序恢复分别管理。

## 真正运行前仍缺少什么

需要实现并锁定 web.react、backend.go、postgres_adapter、go_api_with_static_bundle、ssh_adapter 以及测试运行器。可信 identity provider、应用数据库凭据、受管数据库管理身份、主机用户/指纹、服务账户/目录/sudo 权限和审批状态均需真实绑定。

本示例没有凭空提供用户名密码，没有建立账号系统，没有对外开放端口或生成 TLS。服务默认监听 loopback；公网路由、反向代理、域名、证书、云安全组另行声明。附带测试只有声明，本包没有执行浏览器 E2E 或真实数据库安全测试。

## 验证命令

```bash
python tools/check.py examples/fullstack/project.uicl
```

此命令会读取明确的包内 imports，检查形状、导出和引用。不会联网、建库、SSH、运行模型或部署程序。

## 推荐的真实基准

创建成功、空白标题拒绝、未登录拒绝、跨用户读写拒绝、过期版本冲突、同幂等键不重复创建、部署断线后对账、健康检查失败后仅回退程序版本。把这些真实运行结果与当前结构检查分开记录。
