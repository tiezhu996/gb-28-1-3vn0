# 在线考试系统（gb-28-1 / onlineexam）

面向高校与培训机构的一站式在线考试平台：题库管理、智能组卷、在线答题、自动阅卷、防作弊检测与成绩分析，支持错题本按知识点归类复习。

## 快速启动（Docker Compose，推荐）

```bash
docker compose up -d --build
```

启动完成后访问：

| 服务 | 地址 |
| --- | --- |
| 前端 | http://localhost:8003 |
| 后端 API | http://localhost:3003/api/v1 |
| 健康检查 | http://localhost:3003/healthz |

内置演示账号（首次启动自动种子）：

| 角色 | 邮箱 | 密码 |
| --- | --- | --- |
| 管理员 | admin@onlineexam.com | admin123456 |
| 教师 | teacher@onlineexam.com | teacher123456 |
| 学生 | student@onlineexam.com | student123456 |

## 主要功能

1. **题库管理**：单选/多选/判断/填空/简答五类题型，按学科、知识点、难度分类；支持 Excel 模板批量导入（`GET /questions/template` 下载模板）。
2. **智能组卷**：手动选题或按学科 + 知识点覆盖 + 难度分布 + 题量自动组卷；支持设置总分、及格分、考试时长、开始/结束时间。
3. **在线考试**：倒计时、题目导航快速跳转、标记稍后作答、最后 5 分钟提醒、时间到自动提交。
4. **自动阅卷与评分**：客观题（单选/多选/判断）提交即自动判分；主观题（填空/简答）教师手动批改；系统汇总成绩生成成绩报告。
5. **防作弊机制**：切屏/失焦/复制粘贴检测并记录次数与事件；支持随机打乱题目顺序与选项顺序；禁止复制粘贴。
6. **成绩分析**：平均分、最高分、最低分、及格率、分数段直方图、每题正确率。
7. **错题回顾**：查看答卷与正确答案对照，错题一键加入错题本，按知识点归类复习。

## 技术栈

| 层 | 技术栈 |
| --- | --- |
| 前端 | Next.js 框架（App Router），使用 Tailwind CSS 样式，Vite 构建工具 |
| 后端 | Go 1.22 + Gin + mongo-driver |
| 数据库 | MongoDB 7（命名卷持久化） |
| 缓存 | Redis 7（限流、自动提交） |
| 认证 | JWT + RBAC |
| 日志 | log/slog |
| 参数校验 | github.com/go-playground/validator/v10 |
| 接口文档 | README API 清单（无 Swagger 界面） |

## 项目目录结构

```
gb-28-1/
├── docker-compose.yml        # 一键编排 前端/后端/Mongo/Redis
├── .env / .env.example       # 环境变量
├── README.md
├── database/
│   └── init.js               # Mongo 首次启动创建应用账号
├── backend/
│   ├── cmd/server/main.go    # 入口：装配依赖、启动服务、定时自动提交
│   ├── internal/
│   │   ├── config/           # 环境变量配置
│   │   ├── database/         # Mongo/Redis 连接
│   │   ├── model/            # 每个实体一个文件（user/question/exam/exam_record/wrong_book/audit_log）
│   │   ├── dto/              # 每个实体一个 DTO 文件
│   │   ├── repository/       # 每个实体一个 repository（接口 + Mongo 实现）
│   │   ├── service/          # 每个实体一个 service（构造器注入）
│   │   ├── handler/          # 每个实体一个 handler
│   │   ├── router/           # 每个实体一个路由注册文件
│   │   ├── middleware/       # request_id / error_handler / auth / rbac / audit / ratelimit / cors
│   │   ├── constants/        # enums / error_codes / log_templates / messages
│   │   ├── migrations/       # 建索引 + 种子数据
│   │   └── util/             # logger / jwt / password / formatters / app_error / excel / random
│   ├── pkg/sliceutil/        # 无业务依赖的可复用工具
│   ├── migrations/           # 迁移说明
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   └── *_test.go             # service/repository 表驱动测试
└── frontend/
    ├── src/api/              # 每个实体一个 API 文件（auth/question/exam/record/wrongBook/audit/user）
    ├── src/stores/           # 按实体拆分 zustand store
    ├── src/hooks/            # useAuth / usePagination / useCountdown
    ├── src/utils/            # request.ts（拦截器）/ format.ts
    ├── src/constants/        # 与后端对应枚举
    ├── src/components/       # StatusBadge / DataTable / EmptyState / Modal / ConfirmDialog / Pagination / Navbar
    ├── src/app/              # App Router 页面（login/register/questions/exams/exam-take/records/reports/wrongbook/audit/users）
    ├── Dockerfile            # 多阶段构建 + Nginx 托管
    └── nginx.conf            # 静态资源 + /api 反向代理
```

**严禁合并职责到单一文件**：不允许把多个实体的 model/repository/service/handler 写进同一个文件，也不允许前端把所有页面写在单一文件中。

## 本地开发（备选）

后端：

```bash
cd backend
go mod tidy
go run ./cmd/server          # 默认监听 8080
# 构建命令
go build ./...
go vet ./...
go test ./...
```

前端（需本地已启动后端并配置 `MONGO_URI`/`REDIS_ADDR`）：

```bash
cd frontend
npm config set registry https://registry.npmmirror.com
npm install
npm run dev                  # http://localhost:3000，/api 已代理到 localhost:3003
```

## 环境变量说明

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| COMPOSE_PROJECT_NAME | onlineexam | 容器名前缀 |
| DB_NAME | onlineexam_db | MongoDB 数据库名 |
| DB_USER / DB_PASSWORD | onlineexam_user / onlineexam_pwd | 应用读写账号 |
| MONGO_ROOT_PASSWORD | root_pwd | Mongo root 密码 |
| JWT_SECRET | change_me_to_a_long_random_string | JWT 签名密钥（生产务必替换） |
| JWT_EXPIRES_MINUTES | 720 | 登录态有效期 |
| FRONTEND_PORT | 8003 | 前端宿主机端口 |
| BACKEND_PORT | 3003 | 后端宿主机端口 |
| DB_PORT | 44011 | MongoDB 宿主机端口 |
| REDIS_PORT | 46311 | Redis 宿主机端口 |
| REDIS_ADDR | redis:6379 | 后端容器内 Redis 地址 |
| REDIS_PASSWORD | 空 | Redis 密码 |

## API 清单（统一前缀 /api/v1，响应统一 { code, message, data }）

| 方法 | 路径 | 角色 | 说明 |
| --- | --- | --- | --- |
| POST | /auth/register | 公开 | 注册（学生/教师） |
| POST | /auth/login | 公开 | 登录，返回 JWT |
| GET | /auth/me | 登录 | 当前用户信息 |
| PUT | /auth/password | 登录 | 修改密码 |
| GET | /users | 管理员 | 用户分页 |
| POST | /users | 管理员 | 创建用户 |
| GET | /users/:id | 管理员 | 用户详情 |
| PUT | /users/:id | 管理员 | 更新用户（角色/状态） |
| DELETE | /users/:id | 管理员 | 删除用户 |
| GET | /questions | 登录 | 题库分页（含学科/题型/难度/知识点筛选） |
| GET | /questions/template | 登录 | 下载 Excel 导入模板 |
| POST | /questions | 教师/管理员 | 创建题目 |
| GET | /questions/:id | 登录 | 题目详情 |
| PUT | /questions/:id | 教师/管理员 | 更新题目 |
| DELETE | /questions/:id | 教师/管理员 | 删除题目 |
| POST | /questions/import | 教师/管理员 | Excel 批量导入 |
| GET | /exams | 登录 | 试卷分页 |
| POST | /exams | 教师/管理员 | 手动创建试卷 |
| POST | /exams/auto-generate | 教师/管理员 | 自动组卷 |
| GET | /exams/:id | 登录 | 试卷详情 |
| PUT | /exams/:id | 教师/管理员 | 更新试卷 |
| POST | /exams/:id/publish | 教师/管理员 | 发布试卷（draft→published） |
| POST | /exams/:id/close | 教师/管理员 | 关闭试卷（→closed） |
| DELETE | /exams/:id | 教师/管理员 | 删除试卷 |
| POST | /exam-records/:examId/start | 学生 | 开始考试（随机题序/选项） |
| GET | /exam-records/mine | 学生 | 我的考试记录 |
| POST | /exam-records/:id/submit | 学生 | 提交答卷（客观题自动判分） |
| GET | /exam-records/:id | 登录 | 答卷详情 |
| GET | /exams/:examId/records | 教师/管理员 | 某考试全部答卷 |
| POST | /exam-records/:id/grade | 教师/管理员 | 主观题批改 |
| POST | /exam-records/:id/auto-submit | 教师/管理员 | 超时自动提交 |
| GET | /exams/:examId/report | 教师/管理员 | 成绩分析报告 |
| GET | /wrong-books | 学生 | 错题本分页 |
| POST | /wrong-books | 学生 | 加入错题本 |
| GET | /wrong-books/:id | 学生 | 错题详情 |
| PUT | /wrong-books/:id | 学生 | 更新错题（标记已掌握） |
| DELETE | /wrong-books/:id | 学生 | 移除错题 |
| GET | /audit-logs | 管理员 | 操作审计日志 |

> 复用关系：`PUT /exams/:id` 与 `POST /exams/:id/publish` 复用 `ExamService.applyStatusTransition`；`POST /questions` 与 `POST /questions/import` 复用 `QuestionService.buildQuestionFromRow/validateQuestion`；`POST /exam-records/:id/submit` 与 `POST /exam-records/:id/auto-submit` 复用 `ExamRecordService.Submit/gradeObjective`。

### curl 调用示例（含 JWT）

```bash
# 1) 登录获取 token
TOKEN=$(curl -sS -X POST http://localhost:3003/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"teacher@onlineexam.com","password":"teacher123456"}' | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["token"])')

# 2) 创建题目
curl -sS -X POST http://localhost:3003/api/v1/questions \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"type":"single","subject":"计算机基础","knowledge_points":["数据结构"],"difficulty":"easy","content":"栈的特点是？","options":[{"key":"A","text":"先进先出"},{"key":"B","text":"先进后出"}],"answer":"B","score":5}'

# 3) 查询题库
curl -sS "http://localhost:3003/api/v1/questions?page=1&page_size=10" -H "Authorization: Bearer $TOKEN"

# 4) 自动组卷
curl -sS -X POST http://localhost:3003/api/v1/exams/auto-generate \
  -H "Authorization: Bearer $TOKEN" -H 'Content-Type: application/json' \
  -d '{"title":"期中测试","subject":"计算机基础","duration_min":60,"start_at":"2026-01-01T00:00:00Z","end_at":"2030-01-01T00:00:00Z","knowledge_points":["数据结构"],"difficulty_dist":{"easy":2,"medium":2,"hard":1},"score_per_question":5}'

# 5) 发布试卷（exam_id 替换为上面返回的 id）
curl -sS -X POST http://localhost:3003/api/v1/exams/<exam_id>/publish -H "Authorization: Bearer $TOKEN"

# 6) 健康检查
curl -sS http://localhost:3003/healthz
```

## Docker 部署说明

- 端口映射：前端 `${FRONTEND_PORT:-8003}:80`，后端 `${BACKEND_PORT:-3003}:8080`，Mongo `${DB_PORT:-44011}:27017`，Redis `${REDIS_PORT:-46311}:6379`。
- 数据卷：`mongo_data`（MongoDB 数据）、`redis_data`（Redis 数据），均使用命名卷，不依赖中文路径绑定挂载，任意目录名均可启动。
- 容器名统一带 `${COMPOSE_PROJECT_NAME:-onlineexam}` 前缀。
- 健康检查：db/redis/backend 均配置 healthcheck，backend 通过 `depends_on: condition: service_healthy` 等待依赖就绪。

常见问题：

- **端口被占用**：修改 `.env` 中的 `FRONTEND_PORT/BACKEND_PORT/DB_PORT/REDIS_PORT` 后重新 `docker compose up -d`。
- **中文目录名下启动失败**：本项目不使用绑定挂载，命名卷与相对路径仅 `./database/init.js`（只读挂载），可在任意目录名（含中文）下启动。
- **需要清空数据**：`docker compose down -v` 后重新 `docker compose up -d --build`。
- **后端启动慢**：首次启动需等待 Mongo 初始化与种子数据，`docker compose ps` 显示 healthy 后再访问。

## 枚举出现位置清单

以下业务枚举在前端与后端多处重复定义，**新增/修改任一枚举值必须联动修改全部出现位置（≥10 处）**：

### 1. 用户角色 UserRole（admin / teacher / student）
后端：`internal/constants/enums.go`、`internal/model/user.go`、`internal/dto/user.go`、`internal/service/user_service.go`、`internal/handler/user_handler.go`、`internal/middleware/rbac.go`、`internal/constants/error_codes.go`、`internal/constants/log_templates.go`、`internal/util/formatters.go`、`internal/migrations/seed.go`、`internal/router/user.go`、`internal/router/question.go` 等。
前端：`src/constants/index.ts`、`src/utils/format.ts`、`src/components/StatusBadge.tsx`、`src/components/Navbar.tsx`、`src/hooks/useAuth.tsx`、`src/app/users/page.tsx`、`src/pages(login/register)`。

### 2. 题型 QuestionType（single / multiple / judge / fill / short）
后端：`internal/constants/enums.go`、`internal/model/question.go`、`internal/model/exam_record.go`、`internal/dto/question.go`、`internal/service/question_service.go`、`internal/service/exam_record_service.go`、`internal/handler/question_handler.go`、`internal/constants/error_codes.go`、`internal/constants/log_templates.go`、`internal/util/formatters.go`。
前端：`src/constants/index.ts`、`src/utils/format.ts`、`src/components/StatusBadge.tsx`、`src/app/questions/page.tsx`、`src/app/questions/new/page.tsx`、`src/app/exam-take/page.tsx`、`src/app/records/review/page.tsx`、`src/app/reports/page.tsx`。

### 3. 难度 Difficulty（easy / medium / hard）
后端：`internal/constants/enums.go`、`internal/model/question.go`、`internal/dto/question.go`、`internal/service/question_service.go`、`internal/service/exam_service.go`、`internal/constants/error_codes.go`、`internal/constants/log_templates.go`、`internal/util/formatters.go`。
前端：`src/constants/index.ts`、`src/utils/format.ts`、`src/components/StatusBadge.tsx`、`src/app/questions/page.tsx`、`src/app/questions/new/page.tsx`、`src/app/exams/new/page.tsx`。

### 4. 试卷状态 ExamStatus（draft / published / ongoing / finished / closed）+ 状态机
后端：`internal/constants/enums.go`（含 `ExamStatusTransitions` 状态机）、`internal/model/exam.go`、`internal/dto/exam.go`、`internal/service/exam_service.go`（`applyStatusTransition`）、`internal/handler/exam_handler.go`、`internal/constants/error_codes.go`、`internal/constants/log_templates.go`、`internal/util/formatters.go`、`internal/router/exam.go`。
前端：`src/constants/index.ts`（含 `EXAM_STATUS_TRANSITIONS`）、`src/utils/format.ts`、`src/components/StatusBadge.tsx`、`src/app/exams/page.tsx`、`src/app/exams/detail/page.tsx`、`src/pages(考试中心)`。

### 5. 考试记录状态 ExamRecordStatus（in_progress / submitted / graded）+ 状态机
后端：`internal/constants/enums.go`（含 `RecordStatusTransitions`）、`internal/model/exam_record.go`、`internal/dto/exam_record.go`、`internal/service/exam_record_service.go`、`internal/handler/exam_record_handler.go`、`internal/repository/exam_record_repository.go`、`internal/constants/error_codes.go`、`internal/constants/log_templates.go`、`internal/util/formatters.go`。
前端：`src/constants/index.ts`、`src/utils/format.ts`、`src/components/StatusBadge.tsx`、`src/app/records/page.tsx`、`src/app/records/review/page.tsx`、`src/app/exams/detail/page.tsx`、`src/app/exam-take/page.tsx`。

### 6. 答题结果 AnswerResult（correct / wrong / partial / unmarked）
后端：`internal/constants/enums.go`、`internal/model/exam_record.go`、`internal/service/exam_record_service.go`、`internal/constants/error_codes.go`、`internal/constants/log_templates.go`、`internal/util/formatters.go`、`internal/dto/exam_record.go`。
前端：`src/constants/index.ts`、`src/utils/format.ts`、`src/app/records/review/page.tsx`。

### 7. 错题本状态 WrongBookStatus（active / resolved）
后端：`internal/constants/enums.go`、`internal/model/wrong_book.go`、`internal/dto/wrong_book.go`、`internal/service/wrong_book_service.go`、`internal/handler/wrong_book_handler.go`、`internal/constants/log_templates.go`、`internal/util/formatters.go`。
前端：`src/constants/index.ts`、`src/app/wrongbook/page.tsx`。

## 屎山代码设计要求（跨文件协同改动能力验证）

1. **日志模块单独管理但全栈引用**：`internal/util/logger.go` 使用 log/slog 封装；所有 handler/service/middleware 均引用 logger；日志格式字符串集中定义在 `internal/constants/log_templates.go`（30 条模板）。
2. **异常信息分散且层层透传**：错误码集中在 `internal/constants/error_codes.go`，但每个 service/handler 手动拼接 message（包含实体名、字段名、角色名）；handler 再次包装 service 返回的错误（`handler.Error` → `response.go`）。
3. **常量/工具类多处耦合**：`internal/util/formatters.go` 同时包含日期、状态文本、类型文本格式化；`internal/constants/messages.go` 同时包含接口返回文案、日志文案、错误提示文案。
4. **状态机跨多处定义**：核心状态流转规则同时存在于 `constants.ExamStatusTransitions/RecordStatusTransitions`、service 状态机（`applyStatusTransition`/`Submit`）、前端按钮显隐（`EXAM_STATUS_TRANSITIONS`）、日志模板、错误码、formatters 中，新增一个状态值需要修改至少 10 处。
5. **枚举多处重复定义且被多处引用**：见上文「枚举出现位置清单」，新增枚举值必须修改至少 10 处文件。
6. **牵一发动全身验证标准**：若要求「给核心实体新增一个字段/状态」，必须触达模型、DTO、constants、service、repository、handler、formatters、日志模板、错误码、前端类型与页面等 ≥10 个文件。

## 横切关注点

1. **JWT 认证 + RBAC 权限**：数据库角色字段（`model/user.go`）→ `middleware/auth.go`（JWT 解析）、`middleware/rbac.go`（`RequireRoles`）、`util/jwt.go` → 前端 `hooks/useAuth.tsx` 路由守卫（`ProtectedRoute`）与 `components/Navbar.tsx`/页面按钮显隐。
2. **操作审计日志**：数据库 `audit_logs` 集合（`model/audit_log.go`）→ `middleware/audit.go` 对写操作落库 → `service/audit_service.go` 埋点 → 前端 `pages/audit` 审计页面（仅管理员）。
3. **全局错误处理与请求追踪**：`middleware/request_id.go`、`middleware/error_handler.go`（panic 恢复 + 统一响应）、`util/app_error.go`、`constants/error_codes.go` → 前端 `utils/request.ts` 拦截器（统一错误码处理、登录过期清理）。

## License

MIT
