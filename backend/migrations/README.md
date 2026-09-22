# 数据库迁移说明

本项目使用 MongoDB，采用「启动时自动建索引 + 种子数据」的迁移策略：

- 索引定义：`internal/migrations/indexes.go`（服务启动时自动执行，保证唯一索引与常用查询索引）。
- 种子数据：`internal/migrations/seed.go`（自动创建 admin/teacher/student 三个演示账号）。
- 数据库初始化脚本（容器首次启动创建应用账号）：`../database/init.js`。

如需手工执行 Mongo 脚本：

```bash
mongosh "mongodb://localhost:44011/onlineexam_db" --eval "db.users.createIndex({email:1},{unique:true})"
```
