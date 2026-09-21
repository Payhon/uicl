<!-- uicl:markdown 1.0 -->

<!-- uicl
document #guide_postgresql "PostgreSQL 关系结构与迁移 Profile 1.0"
  version: "1.0"
  status: "consolidated-specification"
  authority: "UICL 项目整合规范；不是外部标准组织认证或完整运行时声明。"
-->

# PostgreSQL 关系结构与迁移 Profile 1.0

标识：`uicl.postgresql@1.0.0`。状态：规范整合稿。依赖：core, backend。

物理关系模式独立于业务模型；不把 SQL 当 Core 表达式。


## `PgColumn`

物理字段。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 列名。 |
| `type` | `text` | 是 | 否 | PostgreSQL 方言类型。 |
| `nullable` | `bool` | 是 | 否 | 可空。 |
| `default` | `value` | 否 | 否 | SQL 字面默认值，不是 SQL 代码。 |
| `default_sql` | `text` | 否 | 否 | 显式 SQL 默认表达式，与 default 互斥。 |

## `PgFK`

外键。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `columns` | `list<text>` | 是 | 否 | 本表列。 |
| `target` | `ref` | 是 | 否 | 目标表。 |
| `target_columns` | `list<text>` | 是 | 否 | 目标列。 |
| `on_delete` | `text` | 是 | 否 | 删除策略。 |

## `PgIndex`

索引。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 索引名。 |
| `columns` | `list<text>` | 是 | 否 | 列名。 |
| `method` | `text` | 否 | 否 | 访问方法。 |
| `unique` | `bool` | 否 | 否 | 唯一索引。 |
| `where_sql` | `text` | 否 | 否 | 显式条件表达式。 |
| `build` | `text` | 否 | 否 | 构建方式。 |

## `PgCheck`

数据库约束。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 约束名。 |
| `sql` | `text` | 是 | 否 | 显式 SQL 表达式；由 PG 适配器解析。 |

## `PgUnique`

唯一键。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `columns` | `list<text>` | 是 | 否 | 联合唯一列。 |

## `pg.database`

受管数据库定义；不包含连接密钥。

主文本：`None`；身份：`required`；允许子节点：pg.table, pg.role, pg.policy, pg.migration, pg.mapping。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 数据库名称。 |
| `engine_major` | `int` | 是 | 否 | 目标数据库主版本。 |
| `schema` | `text` | 否 | 否 | 默认 SQL schema。 |
| `connection` | `ref` | 否 | 否 | 宿主提供的凭据句柄。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

## `pg.table`

物理表，稳定身份与 SQL 名称分离。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 物理表名。 |
| `columns` | `list<record<PgColumn>>` | 是 | 否 | 列声明。 |
| `primary_key` | `list<text>` | 是 | 否 | 联合主键。 |
| `unique` | `list<record<PgUnique>>` | 否 | 否 | 唯一约束。 |
| `foreign_keys` | `list<record<PgFK>>` | 否 | 否 | 外键。 |
| `checks` | `list<record<PgCheck>>` | 否 | 否 | 检查约束。 |
| `indexes` | `list<record<PgIndex>>` | 否 | 否 | 索引。 |
| `rls` | `bool` | 否 | 否 | 启用行级安全。 |
| `force_rls` | `bool` | 否 | 否 | 要求表所有者也受相应策略约束。 |
| `depends_on` | `list<ref>` | 否 | 否 | 显式计划前置依赖；兄弟声明顺序不代表部署顺序。 |

## `pg.role`

数据库角色和执行权限边界。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `name` | `text` | 是 | 否 | 角色名。 |
| `login` | `bool` | 是 | 否 | 是否允许连接。 |
| `bypass_rls` | `bool` | 是 | 否 | 是否绕过 RLS，应用角色必须 false。 |
| `owner` | `bool` | 否 | 否 | 是否承担对象所有权。 |

## `pg.policy`

物理行安全策略。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `table` | `ref` | 是 | 否 | 目标表。 |
| `role` | `ref` | 是 | 否 | 应用角色。 |
| `command` | `text` | 是 | 否 | SQL 操作集合。 |
| `using_sql` | `text` | 否 | 否 | 可见旧行表达式。 |
| `check_sql` | `text` | 否 | 否 | 新行合法性表达式。 |

## `pg.mapping`

业务模型到物理表的显式映射。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `model` | `ref` | 是 | 否 | 模型。 |
| `table` | `ref` | 是 | 否 | 物理表。 |
| `fields` | `record` | 是 | 否 | 逻辑字段到物理列的映射；包含系统字段。 |
| `id_conversion` | `text` | 是 | 否 | 不透明 id 与数据库具体身份类型的转换。 |

## `pg.migration`

有前置版本和显式意图的迁移计划。

主文本：`None`；身份：`required`；允许子节点：无。

| 属性 | 形状类型 | 必需 | 表达式 | 语义 |
|---|---|---|---|---|
| `database` | `ref` | 是 | 否 | 数据库。 |
| `from_version` | `text` | 是 | 否 | 原版本。 |
| `to_version` | `text` | 是 | 否 | 目标版本。 |
| `atomic` | `text` | 是 | 否 | 事务策略。 |
| `destructive` | `text` | 是 | 否 | 破坏性操作门禁。 |
| `steps` | `list<record>` | 是 | 否 | 按注册操作类型解析的迁移步骤；不能直接作为 shell。 |
| `lock_timeout_ms` | `int` | 是 | 否 | 锁等待上限。 |
| `statement_timeout_ms` | `int` | 是 | 否 | 语句超时。 |

输出端口：`plan: MigrationPlan`，可用阶段 `execution`。

## 规则 PG_TYPED_SQL

pg.* 字段类型与 UICL 通用类型分离。sql/default_sql/where_sql 是受方言检查的 SQL，不是 Core = 表达式；SQL 标识符必须正确引用，不拼接不可信文字。


验证类别：`adapter`。

## 规则 PG_TRANSACTION_CONFLICT

包含 concurrent 索引步骤的迁移与 atomic=required 冲突，必须拒绝或取得显式 split_allowed 修订；不能承诺事务外操作自动随事务回滚。


验证类别：`planner`。

## 规则 PG_IDENTITY_AND_RLS

model 策略不自动等同于 RLS。角色、表所有者、BYPASSRLS、会话租户上下文及连接复用必须明确。列改名必须显式步骤，不按相似名称猜测删除重建。


验证类别：`adapter-runtime`。
