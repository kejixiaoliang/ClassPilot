# ClassPilot 第一阶段“班级基础底座”实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 交付一个可在 Windows 10/11 离线运行的班主任工作台，支持当前班级、学生资料、自定义字段、头像附件、导入导出、打印、归档以及完整备份恢复。

**Architecture:** React + TypeScript 构建 Web 界面，Go 单文件程序仅监听 `127.0.0.1` 并提供本地 HTTP API，SQLite 与附件保存在程序目录外部。前端只依赖稳定 API，不直接访问 SQLite、Windows 路径或本地文件系统。

**Tech Stack:** Go 1.26、`net/http`、`modernc.org/sqlite`、`excelize/v2`、React 19.2、TypeScript 7.0、Vite 8.2、浏览器 History API、TanStack Query 5、Zod 4、Vitest 4、Playwright 1.62、SQLite。

## Global Constraints

- 首发平台仅保证 Windows 10/11。
- 浏览器仅保证 Microsoft Edge 和 Google Chrome。
- 核心流程必须在断网环境运行，不加载远程字体、脚本、样式或分析服务。
- 本地服务只能监听 `127.0.0.1`，不得绑定 `0.0.0.0` 或局域网地址。
- 第一版不设置启动密码，也不加密 SQLite 数据库。
- 同一工作台最多存在一个当前班级；历史班级默认只读。
- 前端不得直接访问 SQLite、Windows 绝对路径或本地文件系统。
- 年龄由出生日期动态计算，不持久化为事实字段。
- 数据库与附件只保存相对路径，移动整个工作台目录后必须继续可用。
- 所有数据库结构迁移、批量导入、物理删除和恢复操作执行前必须备份。
- 真实数据库、学生资料、附件、导入导出文件、备份和日志不得进入 Git。
- README、设计文档、实施计划、用户可见文案和 Git 提交信息使用中文。
- 每项任务遵循测试先行；测试失败后才实现最小代码。

---

## 1. 目标文件结构

```text
ClassPilot/
├─ cmd/classpilot/
│  └─ main.go                    # Windows 启动入口
├─ internal/
│  ├─ app/
│  │  ├─ app.go                 # 依赖组装与生命周期
│  │  └─ app_test.go
│  ├─ workspace/
│  │  ├─ workspace.go           # 工作台目录解析与创建
│  │  └─ workspace_test.go
│  ├─ database/
│  │  ├─ database.go            # SQLite 打开、约束、完整性检查
│  │  ├─ migrate.go             # 结构版本迁移
│  │  ├─ database_test.go
│  │  └─ migrations/
│  │     └─ 001_initial.sql
│  ├─ server/
│  │  ├─ server.go              # 回环监听、令牌校验、静态资源
│  │  ├─ response.go            # 统一 JSON 响应与错误
│  │  └─ server_test.go
│  ├─ diagnostics/
│  │  ├─ logger.go              # 脱敏本地日志与错误编号
│  │  └─ logger_test.go
│  ├─ classes/
│  │  ├─ model.go
│  │  ├─ repository.go
│  │  ├─ service.go
│  │  ├─ handler.go
│  │  └─ service_test.go
│  ├─ students/
│  │  ├─ model.go
│  │  ├─ repository.go
│  │  ├─ service.go
│  │  ├─ handler.go
│  │  └─ service_test.go
│  ├─ customfields/
│  │  ├─ model.go
│  │  ├─ repository.go
│  │  ├─ service.go
│  │  ├─ handler.go
│  │  └─ service_test.go
│  ├─ attachments/
│  │  ├─ service.go
│  │  ├─ handler.go
│  │  └─ service_test.go
│  ├─ transfer/
│  │  ├─ model.go
│  │  ├─ parse_csv.go
│  │  ├─ parse_json.go
│  │  ├─ parse_excel.go
│  │  ├─ import_service.go
│  │  ├─ export_service.go
│  │  └─ import_service_test.go
│  ├─ backup/
│  │  ├─ service.go
│  │  ├─ handler.go
│  │  └─ service_test.go
│  └─ audit/
│     ├─ service.go
│     └─ service_test.go
├─ web/
│  ├─ src/
│  │  ├─ app/
│  │  │  ├─ App.tsx
│  │  │  ├─ routes.tsx
│  │  │  └─ queryClient.ts
│  │  ├─ api/
│  │  │  ├─ client.ts
│  │  │  ├─ types.ts
│  │  │  └─ client.test.ts
│  │  ├─ components/
│  │  │  ├─ AppShell.tsx
│  │  │  ├─ DataTable.tsx
│  │  │  ├─ EmptyState.tsx
│  │  │  ├─ ErrorNotice.tsx
│  │  │  └─ ConfirmDialog.tsx
│  │  ├─ features/
│  │  │  ├─ dashboard/
│  │  │  ├─ classes/
│  │  │  ├─ students/
│  │  │  ├─ custom-fields/
│  │  │  ├─ attachments/
│  │  │  ├─ transfer/
│  │  │  ├─ backup/
│  │  │  └─ history/
│  │  ├─ styles/
│  │  │  ├─ tokens.css
│  │  │  ├─ global.css
│  │  │  └─ print.css
│  │  ├─ test/
│  │  │  └─ setup.ts
│  │  └─ main.tsx
│  ├─ e2e/
│  │  └─ phase1.spec.ts
│  ├─ package.json
│  ├─ package-lock.json
│  ├─ vite.config.ts
│  └─ playwright.config.ts
├─ scripts/
│  ├─ build.ps1                 # 构建前端与 Windows 程序
│  └─ smoke-test.ps1            # 发布包冒烟验证
├─ testdata/
│  ├─ students-valid.csv
│  ├─ students-errors.csv
│  └─ students-valid.json
├─ go.mod
├─ go.sum
└─ README.md
```

生产目录由程序首次启动时创建：

```text
ClassPilot/
├─ ClassPilot.exe
├─ data/classpilot.sqlite
├─ attachments/
├─ imports/
├─ exports/
├─ backups/
└─ logs/
```

---

### Task 1: 建立可测试的前后端最小闭环

**Files:**
- Create: `go.mod`
- Create: `cmd/classpilot/main.go`
- Create: `internal/app/app.go`
- Create: `internal/app/app_test.go`
- Create: `web/package.json`
- Create: `web/vite.config.ts`
- Create: `web/tsconfig.json`
- Create: `web/index.html`
- Create: `web/src/main.tsx`
- Create: `web/src/app/App.tsx`
- Create: `web/src/test/setup.ts`
- Modify: `.gitignore`

**Interfaces:**
- Produces: `app.BuildInfo { Version string; Commit string }`
- Produces: `app.New(build BuildInfo) *App`
- Produces: 前端测试、构建和 Go 测试命令。

- [ ] **Step 1: 安装并验证工具链**

Run:

```powershell
winget install --id GoLang.Go --exact
go version
node --version
npm --version
```

Expected: Go 版本不低于 `1.26`，Node 输出 `v22.15.0` 或更高兼容版本，npm 输出 `10.9.2` 或更高兼容版本。

- [ ] **Step 2: 写 Go 入口失败测试**

```go
func TestNewPreservesBuildInfo(t *testing.T) {
    got := app.New(app.BuildInfo{Version: "0.1.0", Commit: "abc123"})
    if got.Build.Version != "0.1.0" || got.Build.Commit != "abc123" {
        t.Fatalf("构建信息未保留：%+v", got.Build)
    }
}
```

Run: `go test ./internal/app -run TestNewPreservesBuildInfo -v`

Expected: FAIL，提示 `app.New` 或 `BuildInfo` 未定义。

- [ ] **Step 3: 实现最小 Go 应用对象**

```go
type BuildInfo struct {
    Version string
    Commit  string
}

type App struct {
    Build BuildInfo
}

func New(build BuildInfo) *App {
    return &App{Build: build}
}
```

Run: `go test ./internal/app -v`

Expected: PASS。

- [ ] **Step 4: 建立 React 测试与构建入口**

`web/src/app/App.tsx`：

```tsx
export function App() {
  return <main><h1>ClassPilot 班主任工作台</h1></main>;
}
```

`web/src/app/App.test.tsx`：

```tsx
it("显示中文产品名称", () => {
  render(<App />);
  expect(screen.getByRole("heading", { name: "ClassPilot 班主任工作台" })).toBeInTheDocument();
});
```

Run:

```powershell
Set-Location web
npm install react@19.2.8 react-dom@19.2.8 "@tanstack/react-query@5.101.4" zod@4.4.3
npm install --save-dev typescript@7.0.2 vite@8.2.0 vitest@4.1.10 "@playwright/test@1.62.0" "@vitejs/plugin-react" "@testing-library/react" "@testing-library/jest-dom" jsdom
npm test -- --run
npm run build
```

Expected: 测试与构建均 PASS，生成 `web/dist/`。

- [ ] **Step 5: 验证全仓库基础命令**

Run:

```powershell
go test ./...
npm --prefix web test -- --run
npm --prefix web run build
```

Expected: 全部 PASS。

安全说明：React Router 在计划执行时存在尚未覆盖完整安全公告范围的版本冲突，而且第一阶段不需要服务端路由能力，因此改用浏览器 History API 的小型本地路由封装，并保持 `npm audit` 为零漏洞。

- [ ] **Step 6: 中文提交**

```powershell
git add go.mod cmd internal/app web .gitignore
git commit -m "工程：建立前后端测试与构建基础"
```

---

### Task 2: 工作台目录与 SQLite 初始结构

**Files:**
- Create: `internal/workspace/workspace.go`
- Create: `internal/workspace/workspace_test.go`
- Create: `internal/database/database.go`
- Create: `internal/database/migrate.go`
- Create: `internal/database/database_test.go`
- Create: `internal/database/migrations/001_initial.sql`
- Modify: `go.mod`
- Modify: `go.sum`

**Interfaces:**
- Produces: `workspace.Paths`
- Produces: `workspace.Ensure(root string) (Paths, error)`
- Produces: `workspace.AcquireLock(root string) (release func() error, err error)`
- Produces: `database.Open(path string) (*sql.DB, error)`
- Produces: `database.Migrate(ctx context.Context, db *sql.DB) error`
- Produces: `database.IntegrityCheck(ctx context.Context, db *sql.DB) error`
- Produces: `database.Backup(ctx context.Context, db *sql.DB, destination string) error`

- [ ] **Step 1: 写工作台目录失败测试**

```go
func TestEnsureCreatesPortableWorkspace(t *testing.T) {
    root := t.TempDir()
    paths, err := workspace.Ensure(root)
    if err != nil { t.Fatal(err) }
    for _, dir := range []string{paths.Data, paths.Attachments, paths.Imports, paths.Exports, paths.Backups, paths.Logs} {
        if info, err := os.Stat(dir); err != nil || !info.IsDir() {
            t.Fatalf("目录未创建：%s", dir)
        }
    }
    if filepath.IsAbs(paths.DatabaseRelative) {
        t.Fatalf("数据库业务路径必须是相对路径：%s", paths.DatabaseRelative)
    }
}

func TestAcquireLockRejectsSecondWriter(t *testing.T) {
    root := t.TempDir()
    release, err := workspace.AcquireLock(root)
    if err != nil { t.Fatal(err) }
    defer release()
    if _, err := workspace.AcquireLock(root); err == nil {
        t.Fatal("第二个写入实例应被拒绝")
    }
}
```

Run: `go test ./internal/workspace -v`

Expected: FAIL，提示 `workspace.Ensure` 未定义。

- [ ] **Step 2: 实现目录解析**

```go
type Paths struct {
    Root, Data, Attachments, Imports, Exports, Backups, Logs string
    Database, DatabaseRelative string
}

func Ensure(root string) (Paths, error) {
    // 将 root 转为绝对路径，创建六个业务目录；
    // DatabaseRelative 固定为 filepath.Join("data", "classpilot.sqlite")。
}

func AcquireLock(root string) (func() error, error)
```

锁文件只保存进程标识和启动时间；程序正常退出时释放。发现陈旧锁时，先确认对应进程不存在再接管。

Run: `go test ./internal/workspace -v`

Expected: PASS。

- [ ] **Step 3: 写数据库迁移失败测试**

```go
func TestMigrateCreatesInitialSchema(t *testing.T) {
    db := openTestDB(t)
    if err := database.Migrate(context.Background(), db); err != nil { t.Fatal(err) }
    for _, table := range []string{"schema_migrations", "classes", "students", "custom_fields", "student_field_values", "attachments", "audit_logs"} {
        assertTableExists(t, db, table)
    }
}
```

Run: `go test ./internal/database -run TestMigrateCreatesInitialSchema -v`

Expected: FAIL，提示迁移函数或表不存在。

- [ ] **Step 4: 编写初始 SQL 结构**

关键约束必须直接写入 `001_initial.sql`：

```sql
CREATE UNIQUE INDEX one_current_class
ON classes(status)
WHERE status = 'current';

CREATE UNIQUE INDEX unique_student_number_per_class
ON students(class_id, student_number)
WHERE student_number IS NOT NULL AND student_number <> '' AND deleted_at IS NULL;

PRAGMA foreign_keys = ON;
```

所有业务表使用文本 UUID 主键和 ISO 8601 UTC 时间；`students` 包含 `deleted_at`，附件路径列命名为 `relative_path`。

- [ ] **Step 5: 实现 SQLite 打开、迁移和完整性检查**

```go
func Open(path string) (*sql.DB, error)
func Migrate(ctx context.Context, db *sql.DB) error
func IntegrityCheck(ctx context.Context, db *sql.DB) error
func Backup(ctx context.Context, db *sql.DB, destination string) error
```

`Open` 必须启用 `foreign_keys`、`busy_timeout=5000`，并限制为单写连接；`IntegrityCheck` 执行 `PRAGMA integrity_check` 并要求结果为 `ok`；`Backup` 使用 SQLite 在线备份能力写入临时文件，完整性检查通过后再原子改名。

Run: `go test ./internal/database -v`

Expected: PASS。

- [ ] **Step 6: 中文提交**

```powershell
git add go.mod go.sum internal/workspace internal/database
git commit -m "数据：建立工作台目录与初始数据库结构"
```

---

### Task 3: 本地服务、访问令牌与统一错误响应

**Files:**
- Create: `internal/server/server.go`
- Create: `internal/server/response.go`
- Create: `internal/server/server_test.go`
- Create: `internal/audit/service.go`
- Create: `internal/audit/service_test.go`
- Create: `internal/diagnostics/logger.go`
- Create: `internal/diagnostics/logger_test.go`
- Modify: `internal/app/app.go`
- Modify: `cmd/classpilot/main.go`

**Interfaces:**
- Produces: `server.Config { Token string; Static http.FileSystem; ReadOnly bool }`
- Produces: `server.New(config Config) http.Handler`
- Produces: `server.ListenLoopback() (net.Listener, error)`
- Produces: `server.APIError { Code string; Message string; Details any }`
- Produces: `GET /api/v1/health`
- Produces: `audit.Entry`
- Produces: `audit.Service.Record(ctx context.Context, entry Entry) error`
- Produces: `diagnostics.Logger.Error(code string, err error, fields map[string]any)`

- [ ] **Step 1: 写回环监听与令牌失败测试**

```go
func TestAPIRejectsMissingToken(t *testing.T) {
    handler := server.New(server.Config{Token: "secret"})
    req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
    res := httptest.NewRecorder()
    handler.ServeHTTP(res, req)
    if res.Code != http.StatusUnauthorized { t.Fatalf("状态码=%d", res.Code) }
}

func TestAPIAcceptsValidToken(t *testing.T) {
    handler := server.New(server.Config{Token: "secret"})
    req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
    req.Header.Set("X-ClassPilot-Token", "secret")
    res := httptest.NewRecorder()
    handler.ServeHTTP(res, req)
    if res.Code != http.StatusOK { t.Fatalf("状态码=%d", res.Code) }
}
```

Run: `go test ./internal/server -v`

Expected: FAIL。

- [ ] **Step 2: 实现统一响应和令牌中间件**

```go
type APIError struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Details any    `json:"details,omitempty"`
}

type Envelope[T any] struct {
    Data  *T        `json:"data,omitempty"`
    Error *APIError `json:"error,omitempty"`
}
```

错误必须返回稳定中文信息与机器可读 `code`，不得把 SQL、绝对路径或堆栈发送到浏览器。

- [ ] **Step 3: 实现只监听回环地址**

```go
func ListenLoopback() (net.Listener, error) {
    return net.Listen("tcp4", "127.0.0.1:0")
}
```

增加测试断言 `listener.Addr()` 的 IP 是回环地址。

- [ ] **Step 4: 组装启动入口**

```go
func main() {
    // 解析 exe 所在目录为工作台根目录；
    // 创建随机 32 字节令牌；
    // 启动本地服务；
    // 使用带片段令牌的 URL 打开默认浏览器；
    // 接收 Ctrl+C 后优雅关闭。
}
```

启动顺序固定为：获取工作台写锁、打开数据库、执行迁移前备份、执行迁移、完整性检查、启动 HTTP 服务、打开浏览器。完整性检查失败时以 `ReadOnly: true` 启动保护模式，所有写请求返回 `DATABASE_READ_ONLY`。令牌通过 URL fragment 传递，前端读取后保存到 `sessionStorage` 并立即从地址栏移除，避免进入 HTTP 日志。

- [ ] **Step 5: 实现脱敏操作记录基础服务**

```go
type Entry struct {
    ID, Action, EntityType, EntityID, Summary string
    CreatedAt time.Time
}

func (s *Service) Record(ctx context.Context, entry Entry) error
```

测试必须断言摘要中出现身份证号、电话号码或绝对路径时拒绝写入，并返回 `SENSITIVE_AUDIT_CONTENT`。

- [ ] **Step 6: 实现脱敏本地日志**

```go
type Logger struct {
    output io.Writer
}

func (l *Logger) Error(code string, err error, fields map[string]any)
```

先写 `TestLoggerRedactsPhoneIDCardAndAbsolutePath`，断言日志只包含错误编号、错误类型和脱敏上下文，不包含原始电话、身份证号、学生字段值、附件内容或工作台绝对路径。日志使用 JSON Lines 写入 `logs/classpilot.log`。

- [ ] **Step 7: 验证**

Run:

```powershell
go test ./internal/server ./internal/app ./internal/audit ./internal/diagnostics -v
go test ./...
```

Expected: PASS，测试监听地址均为 `127.0.0.1`。

- [ ] **Step 8: 中文提交**

```powershell
git add cmd/classpilot internal/app internal/server internal/audit internal/diagnostics
git commit -m "服务：建立回环接口与临时访问令牌"
```

---

### Task 4: 当前班级、归档与历史班级 API

**Files:**
- Create: `internal/classes/model.go`
- Create: `internal/classes/repository.go`
- Create: `internal/classes/service.go`
- Create: `internal/classes/handler.go`
- Create: `internal/classes/service_test.go`
- Modify: `internal/server/server.go`
- Modify: `internal/server/server_test.go`

**Interfaces:**
- Consumes: `audit.Service.Record`
- Produces: `classes.Class`
- Produces: `classes.CreateInput`
- Produces: `classes.UpdateInput`
- Produces: `classes.Service.CreateCurrent(ctx, input) (Class, error)`
- Produces: `classes.Service.GetCurrent(ctx) (*Class, error)`
- Produces: `classes.Service.ListArchived(ctx) ([]Class, error)`
- Produces: `classes.Service.ArchiveCurrent(ctx, id string) error`
- Produces: `classes.Service.RestoreAsCurrent(ctx, id string) error`
- Produces: `/api/v1/classes/current` 与 `/api/v1/classes/archived`

- [ ] **Step 1: 写唯一当前班级失败测试**

```go
func TestOnlyOneCurrentClass(t *testing.T) {
    svc := newTestService(t)
    _, err := svc.CreateCurrent(ctx, classes.CreateInput{Name: "七年级一班", Stage: "初中", Grade: "七年级", SchoolYear: "2026-2027"})
    if err != nil { t.Fatal(err) }
    _, err = svc.CreateCurrent(ctx, classes.CreateInput{Name: "七年级二班", Stage: "初中", Grade: "七年级", SchoolYear: "2026-2027"})
    assertDomainCode(t, err, "CURRENT_CLASS_EXISTS")
}
```

Run: `go test ./internal/classes -run TestOnlyOneCurrentClass -v`

Expected: FAIL。

- [ ] **Step 2: 实现班级模型与仓储**

```go
type Class struct {
    ID, Name, Stage, Grade, SchoolYear, Status, Notes string
    CreatedAt, UpdatedAt time.Time
    ArchivedAt *time.Time
}
```

仓储方法仅负责 SQL；业务服务负责唯一当前班级、归档只读和恢复冲突规则。

- [ ] **Step 3: 写归档与恢复冲突测试**

```go
func TestArchivedClassIsReadOnly(t *testing.T)
func TestRestoreRequiresNoCurrentClass(t *testing.T)
```

Run: `go test ./internal/classes -v`

Expected: 新测试先 FAIL。

- [ ] **Step 4: 实现归档、恢复和 HTTP 处理器**

REST 契约：

```text
GET    /api/v1/classes/current
POST   /api/v1/classes/current
PATCH  /api/v1/classes/current
POST   /api/v1/classes/{id}/archive
GET    /api/v1/classes/archived
POST   /api/v1/classes/{id}/restore
```

归档和恢复必须写入 `audit_logs`。

- [ ] **Step 5: 验证**

Run: `go test ./internal/classes ./internal/server -v`

Expected: PASS。

- [ ] **Step 6: 中文提交**

```powershell
git add internal/classes internal/server
git commit -m "班级：实现当前班级与历史归档"
```

---

### Task 5: 学生资料、查询与回收站 API

**Files:**
- Create: `internal/students/model.go`
- Create: `internal/students/repository.go`
- Create: `internal/students/service.go`
- Create: `internal/students/handler.go`
- Create: `internal/students/service_test.go`
- Modify: `internal/server/server.go`

**Interfaces:**
- Produces: `students.Student`
- Produces: `students.CreateInput`、`students.UpdateInput`、`students.Query`
- Produces: `students.Filter`、`students.BatchUpdateInput`
- Produces: `students.Page { Items []Student; Total int; Page int; PageSize int }`
- Produces: `students.Service.Create/Get/Update/BatchUpdate/List/Trash/Restore`
- Produces: `/api/v1/students`

- [ ] **Step 1: 写学生约束失败测试**

```go
func TestStudentNumberUniqueWithinClass(t *testing.T)
func TestSameStudentNumberAllowedAcrossClasses(t *testing.T)
func TestAgeIsCalculatedFromBirthDate(t *testing.T)
func TestArchivedClassRejectsStudentWrites(t *testing.T)
```

`TestAgeIsCalculatedFromBirthDate` 使用固定计算日期，禁止依赖当前系统时间。

Run: `go test ./internal/students -v`

Expected: FAIL。

- [ ] **Step 2: 实现模型和校验**

```go
type Student struct {
    ID, ClassID, Name, Gender, StudentNumber, Notes string
    BirthDate, EnrollmentDate *time.Time
    CreatedAt, UpdatedAt time.Time
    DeletedAt *time.Time
}

func AgeAt(birthDate, at time.Time) int
```

姓名必填；性别允许使用配置外文本但前端默认提供“男、女、未填写”；日期统一按本地日历日期处理。

- [ ] **Step 3: 写查询和回收站失败测试**

```go
func TestListExcludesTrashedByDefault(t *testing.T)
func TestListFiltersNameAndStudentNumber(t *testing.T)
func TestListCombinesCoreAndCustomFieldFilters(t *testing.T)
func TestBatchUpdateRollsBackAllStudentsOnValidationError(t *testing.T)
func TestTrashThenRestoreStudent(t *testing.T)
```

Run: `go test ./internal/students -v`

Expected: 新测试先 FAIL。

- [ ] **Step 4: 实现分页、筛选、排序和删除语义**

```go
type Query struct {
    ClassID, Search, SortBy, SortOrder string
    Page, PageSize int
    IncludeTrashed bool
    Filters []Filter
}

type Filter struct {
    FieldID string
    Operator string
    Value any
}
```

筛选操作符白名单为 `equals`、`contains`、`in`、`before`、`after`、`is_empty`；服务端根据字段类型拒绝不适用的操作符。`SortBy` 只允许白名单字段。批量编辑只允许备注、入班日期、状态和已启用自定义字段，并在单个事务内完成。第一阶段的物理删除在 Task 11 接入完整备份后实现，本任务只实现可恢复的逻辑删除。

- [ ] **Step 5: 实现 HTTP API**

```text
GET    /api/v1/students
POST   /api/v1/students
GET    /api/v1/students/{id}
PATCH  /api/v1/students/{id}
PATCH  /api/v1/students/batch
POST   /api/v1/students/{id}/trash
POST   /api/v1/students/{id}/restore
```

- [ ] **Step 6: 验证并提交**

Run: `go test ./internal/students ./internal/classes ./internal/server -v`

Expected: PASS。

```powershell
git add internal/students internal/server
git commit -m "学生：实现资料管理与回收站"
```

---

### Task 6: 自定义字段与学生字段值

**Files:**
- Create: `internal/customfields/model.go`
- Create: `internal/customfields/repository.go`
- Create: `internal/customfields/service.go`
- Create: `internal/customfields/handler.go`
- Create: `internal/customfields/service_test.go`
- Modify: `internal/students/model.go`
- Modify: `internal/students/service.go`
- Modify: `internal/server/server.go`

**Interfaces:**
- Consumes: `audit.Service.Record`
- Produces: `customfields.FieldType`
- Produces: `customfields.FieldDefinition`
- Produces: `customfields.Value`
- Produces: `customfields.Service.Create/Update/Disable/Reorder/ValidateValues`
- Extends: `students.CreateInput.CustomValues map[string]any`
- Extends: `students.UpdateInput.CustomValues map[string]any`

- [ ] **Step 1: 写字段类型校验失败测试**

```go
func TestValidateSupportedFieldTypes(t *testing.T)
func TestRequiredFieldRejectsEmptyValue(t *testing.T)
func TestSingleChoiceRejectsUnknownOption(t *testing.T)
func TestDisabledFieldPreservesExistingValue(t *testing.T)
```

支持类型固定为：`text`、`long_text`、`number`、`date`、`single_choice`、`multi_choice`、`boolean`。

Run: `go test ./internal/customfields -v`

Expected: FAIL。

- [ ] **Step 2: 实现字段定义和值校验**

```go
type FieldDefinition struct {
    ID, ClassID, Name string
    Type FieldType
    Required, Enabled, ShowInList bool
    Options []string
    SortOrder int
}

func (s *Service) ValidateValues(ctx context.Context, classID string, values map[string]any) (map[string]string, error)
```

值存储前规范化为 JSON；数字不能用浮点字符串蒙混通过，日期必须为 `YYYY-MM-DD`。

- [ ] **Step 3: 写学生与自定义字段事务测试**

```go
func TestStudentAndCustomValuesSaveAtomically(t *testing.T)
```

故意提供一个非法字段值，断言学生和字段值均未写入。

- [ ] **Step 4: 实现 API 与学生服务集成**

```text
GET    /api/v1/custom-fields
POST   /api/v1/custom-fields
PATCH  /api/v1/custom-fields/{id}
POST   /api/v1/custom-fields/{id}/disable
PUT    /api/v1/custom-fields/order
```

- [ ] **Step 5: 验证并提交**

Run: `go test ./internal/customfields ./internal/students -v`

Expected: PASS。

```powershell
git add internal/customfields internal/students internal/server
git commit -m "字段：支持班级自定义学生字段"
```

---

### Task 7: Web 应用外壳、当前班级与学生管理界面

**Files:**
- Create: `web/src/app/routes.tsx`
- Create: `web/src/app/queryClient.ts`
- Create: `web/src/api/client.ts`
- Create: `web/src/api/types.ts`
- Create: `web/src/api/client.test.ts`
- Create: `web/src/components/AppShell.tsx`
- Create: `web/src/components/DataTable.tsx`
- Create: `web/src/components/EmptyState.tsx`
- Create: `web/src/components/ErrorNotice.tsx`
- Create: `web/src/components/ConfirmDialog.tsx`
- Create: `web/src/features/dashboard/DashboardPage.tsx`
- Create: `web/src/features/classes/ClassOverviewPage.tsx`
- Create: `web/src/features/students/StudentListPage.tsx`
- Create: `web/src/features/students/StudentDetailPage.tsx`
- Create: `web/src/features/students/StudentForm.tsx`
- Create: `web/src/features/custom-fields/CustomFieldsPage.tsx`
- Create: `web/src/features/settings/SettingsPage.tsx`
- Create: `web/src/styles/tokens.css`
- Create: `web/src/styles/global.css`
- Modify: `web/src/app/App.tsx`

**Interfaces:**
- Consumes: Task 3–6 的 `/api/v1` 契约。
- Produces: `api.get/post/patch/delete`
- Produces: 当前班级、学生、自定义字段页面路由。

- [ ] **Step 1: 写令牌处理与错误映射失败测试**

```tsx
it("从地址片段读取令牌并立即移除", () => {
  window.location.hash = "token=abc";
  bootstrapToken();
  expect(sessionStorage.getItem("classpilot-token")).toBe("abc");
  expect(window.location.hash).toBe("");
});

it("把 API 错误映射为中文可展示对象", async () => {
  // mock fetch 返回 { error: { code: "STUDENT_NUMBER_EXISTS", message: "学号已存在" } }
  await expect(api.get("/students")).rejects.toMatchObject({ code: "STUDENT_NUMBER_EXISTS" });
});
```

Run: `npm --prefix web test -- --run src/api/client.test.ts`

Expected: FAIL。

- [ ] **Step 2: 实现 API 客户端和查询缓存**

```ts
export type ApiError = { code: string; message: string; details?: unknown };
export const api = {
  get: <T>(path: string) => request<T>("GET", path),
  post: <T>(path: string, body?: unknown) => request<T>("POST", path, body),
  patch: <T>(path: string, body: unknown) => request<T>("PATCH", path, body),
  delete: <T>(path: string) => request<T>("DELETE", path),
};
```

所有请求添加 `X-ClassPilot-Token`；401 显示“工作台会话已失效，请重新启动”。

- [ ] **Step 3: 写导航与空状态失败测试**

```tsx
it("没有当前班级时引导创建班级", async () => {
  renderAppAt("/");
  expect(await screen.findByText("创建第一个班级")).toBeInTheDocument();
});

it("侧栏包含第一阶段页面", () => {
  expect(screen.getByRole("link", { name: "学生名单" })).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "数据中心" })).toBeInTheDocument();
  expect(screen.getByRole("link", { name: "历史班级" })).toBeInTheDocument();
});
```

- [ ] **Step 4: 实现响应式应用外壳**

使用 CSS 自定义属性建立颜色、间距、圆角、阴影和排版变量。不得依赖在线字体；优先使用 Windows 系统中文字体栈。

导航路由：

```text
/
/class
/students
/students/:id
/custom-fields
/data
/history
/settings
```

- [ ] **Step 5: 实现班级和学生页面**

学生名单提供表格／卡片切换、搜索、核心字段与自定义字段组合筛选、排序、分页、显示列、批量编辑和回收站入口；学生表单复用新增和编辑模式；敏感信息默认显示掩码。

- [ ] **Step 6: 实现自定义字段设置页面**

字段创建表单根据类型显示选项设置；停用字段必须展示“已有数据将保留”的确认文案。

设置页面第一阶段只显示应用版本、工作台目录、数据库位置、附件位置和支持范围；本地绝对路径只在用户主动展开后显示。模块启停在对应业务模块进入实施阶段后增加，不创建无效开关。

- [ ] **Step 7: 验证并提交**

Run:

```powershell
npm --prefix web test -- --run
npm --prefix web run build
```

Expected: PASS，无 TypeScript 错误。

```powershell
git add web
git commit -m "界面：完成班级与学生基础管理页面"
```

---

### Task 8: Excel、JSON、CSV 导入预览与事务提交

**Files:**
- Create: `internal/transfer/model.go`
- Create: `internal/transfer/parse_csv.go`
- Create: `internal/transfer/parse_json.go`
- Create: `internal/transfer/parse_excel.go`
- Create: `internal/transfer/import_service.go`
- Create: `internal/transfer/import_service_test.go`
- Create: `testdata/students-valid.csv`
- Create: `testdata/students-errors.csv`
- Create: `testdata/students-valid.json`
- Create: `web/src/features/transfer/ImportPage.tsx`
- Create: `web/src/features/transfer/ImportPage.test.tsx`
- Modify: `internal/server/server.go`
- Modify: `web/src/app/routes.tsx`

**Interfaces:**
- Consumes: `database.Backup`
- Consumes: `audit.Service.Record`
- Produces: `transfer.Preview`
- Produces: `transfer.ColumnMapping`
- Produces: `transfer.ImportDecision`
- Produces: `transfer.Service.Preview(ctx, file, format) (Preview, error)`
- Produces: `transfer.Service.Commit(ctx, previewID string, mappings []ColumnMapping, decisions []ImportDecision) (ImportResult, error)`
- Produces: `/api/v1/imports/preview` 与 `/api/v1/imports/commit`

- [ ] **Step 1: 写三种格式解析失败测试**

```go
func TestPreviewCSVPreservesChineseHeaders(t *testing.T)
func TestPreviewJSONDistinguishesBusinessDataFromBackup(t *testing.T)
func TestPreviewExcelUsesFirstNonEmptySheet(t *testing.T)
```

Run: `go test ./internal/transfer -run TestPreview -v`

Expected: FAIL。

- [ ] **Step 2: 实现解析器**

解析器统一输出：

```go
type RawTable struct {
    Headers []string
    Rows    []map[string]string
}
```

CSV 自动识别 UTF-8 BOM；不猜测无法确定的编码，而是返回 `UNSUPPORTED_TEXT_ENCODING`。Excel 仅接受 `.xlsx`；JSON 必须校验明确的 `kind` 和 `version`。

- [ ] **Step 3: 写校验、冲突和回滚失败测试**

```go
func TestPreviewSeparatesValidWarningsAndErrors(t *testing.T)
func TestDuplicateStudentOffersSkipUpdateCreate(t *testing.T)
func TestCommitRollsBackEveryRowOnFailure(t *testing.T)
func TestCommitRequiresSuccessfulPreImportBackup(t *testing.T)
```

- [ ] **Step 4: 实现预览缓存与事务提交**

预览结果使用随机 `previewID` 保存在内存，15 分钟后失效；提交时重新校验文件摘要、映射和班级状态。提交顺序固定为：调用 `database.Backup` 创建导入前备份、开启事务、写学生与字段值、写审计、提交事务。

- [ ] **Step 5: 实现三步导入页面**

```text
选择文件 → 字段映射与错误预览 → 冲突决策与提交结果
```

错误表必须显示原始行号、字段、原值和中文原因，并支持下载 CSV 错误报告。

- [ ] **Step 6: 验证并提交**

Run:

```powershell
go test ./internal/transfer -v
npm --prefix web test -- --run src/features/transfer
```

Expected: PASS。

```powershell
git add internal/transfer internal/server testdata web/src/features/transfer web/src/app/routes.tsx go.mod go.sum
git commit -m "导入：支持名单预览校验与事务写入"
```

---

### Task 9: 学生头像与附件

**Files:**
- Create: `internal/attachments/service.go`
- Create: `internal/attachments/handler.go`
- Create: `internal/attachments/service_test.go`
- Create: `web/src/features/attachments/AttachmentPanel.tsx`
- Create: `web/src/features/attachments/AttachmentPanel.test.tsx`
- Modify: `internal/students/handler.go`
- Modify: `internal/server/server.go`
- Modify: `web/src/features/students/StudentDetailPage.tsx`

**Interfaces:**
- Produces: `attachments.Metadata`
- Produces: `attachments.Service.Upload/Open/Delete/List`
- Produces: `/api/v1/students/{id}/attachments`
- Produces: `/api/v1/attachments/{id}/content`

- [ ] **Step 1: 写安全路径与文件校验失败测试**

```go
func TestUploadCopiesFileUnderWorkspace(t *testing.T)
func TestUploadStoresRelativePathOnly(t *testing.T)
func TestOpenRejectsPathTraversal(t *testing.T)
func TestUploadRejectsOversizedFile(t *testing.T)
func TestUploadCalculatesSHA256(t *testing.T)
```

默认限制：头像 10 MiB，普通附件 50 MiB；允许的头像类型为 JPEG、PNG、WebP；普通附件允许 PDF、Office 文档和常见图片。

- [ ] **Step 2: 实现附件服务**

```go
type Metadata struct {
    ID, StudentID, Category, OriginalName, RelativePath, MIMEType, SHA256 string
    Size int64
    Description string
    CreatedAt time.Time
}
```

目标文件名使用 UUID 与经过清洗的扩展名；先写临时文件并校验，再原子移动到正式位置。数据库写入失败时删除临时文件。

- [ ] **Step 3: 实现附件 API 和前端面板**

附件查看通过受令牌保护的 API 流式返回，不暴露本地路径。删除附件先移入附件回收目录并写审计，待完整备份后才能物理清理。

- [ ] **Step 4: 验证并提交**

Run:

```powershell
go test ./internal/attachments -v
npm --prefix web test -- --run src/features/attachments
```

Expected: PASS。

```powershell
git add internal/attachments internal/server internal/students web/src/features/attachments web/src/features/students
git commit -m "附件：支持学生头像与资料文件"
```

---

### Task 10: 数据导出与基础名单打印

**Files:**
- Create: `internal/transfer/export_service.go`
- Create: `internal/transfer/export_service_test.go`
- Create: `web/src/features/transfer/ExportPage.tsx`
- Create: `web/src/features/transfer/ExportPage.test.tsx`
- Create: `web/src/features/students/PrintStudentListPage.tsx`
- Create: `web/src/styles/print.css`
- Modify: `internal/server/server.go`
- Modify: `web/src/app/routes.tsx`

**Interfaces:**
- Produces: `transfer.ExportRequest`
- Produces: `transfer.ExportResult`
- Produces: `transfer.Service.ExportCSV/ExportJSON/ExportExcel`
- Produces: `POST /api/v1/exports`
- Produces: `/students/print`

- [ ] **Step 1: 写导出内容失败测试**

```go
func TestCSVExportUsesUTF8BOMAndSelectedFields(t *testing.T)
func TestJSONExportIncludesFormatVersion(t *testing.T)
func TestExcelExportPreservesChineseAndDates(t *testing.T)
func TestExportExcludesTrashedStudentsByDefault(t *testing.T)
```

Run: `go test ./internal/transfer -run Export -v`

Expected: FAIL。

- [ ] **Step 2: 实现业务导出**

```go
type ExportRequest struct {
    ClassID string
    Format  string
    FieldIDs []string
    IncludeTrashed bool
}
```

生成文件写入 `exports/`，API 返回文件标识和下载地址，不返回绝对路径。文件名清洗 Windows 禁止字符并包含班级与生成时间。

- [ ] **Step 3: 写打印页面失败测试**

```tsx
it("只打印用户选择的列", async () => {
  renderPrintPage({ columns: ["姓名", "学号"] });
  expect(await screen.findByRole("columnheader", { name: "姓名" })).toBeInTheDocument();
  expect(screen.queryByRole("columnheader", { name: "家庭电话" })).not.toBeInTheDocument();
});
```

- [ ] **Step 4: 实现导出与打印界面**

打印设置支持标题、显示列、纵向／横向和打印日期。`print.css` 隐藏导航与按钮，使用毫米单位控制 A4 边距，不远程加载字体。

- [ ] **Step 5: 验证并提交**

Run:

```powershell
go test ./internal/transfer -v
npm --prefix web test -- --run src/features/transfer src/features/students
```

Expected: PASS。

```powershell
git add internal/transfer internal/server web/src/features/transfer web/src/features/students web/src/styles web/src/app/routes.tsx
git commit -m "导出：实现名单数据导出与基础打印"
```

---

### Task 11: 操作记录、完整备份与安全恢复

**Files:**
- Create: `internal/backup/service.go`
- Create: `internal/backup/handler.go`
- Create: `internal/backup/service_test.go`
- Create: `web/src/features/backup/BackupPage.tsx`
- Create: `web/src/features/backup/BackupPage.test.tsx`
- Create: `web/src/features/backup/AuditLogPage.tsx`
- Modify: `internal/audit/service.go`
- Modify: `internal/audit/service_test.go`
- Modify: `internal/students/service.go`
- Modify: `internal/students/handler.go`
- Modify: `internal/students/service_test.go`
- Modify: `internal/server/server.go`
- Modify: `web/src/app/routes.tsx`

**Interfaces:**
- Consumes: `audit.Entry`
- Extends: `audit.Service.List`
- Produces: `backup.Manifest`
- Produces: `backup.Service.CreateDatabaseBackup/CreateWorkspaceBackup/Validate/Restore`
- Produces: `students.Service.Purge(ctx context.Context, studentID string) error`
- Produces: `/api/v1/backups`、`/api/v1/restores`、`/api/v1/audit-logs`

- [ ] **Step 1: 写脱敏审计失败测试**

```go
func TestAuditDoesNotStoreSensitiveValues(t *testing.T)
func TestAuditRecordsImportDeleteArchiveAndRestore(t *testing.T)
```

审计记录只保存操作类型、对象类型、对象标识、摘要和时间，不保存身份证号、电话、附件内容或整行导入数据。

- [ ] **Step 2: 写备份完整性失败测试**

```go
func TestDatabaseBackupPassesIntegrityCheck(t *testing.T)
func TestWorkspaceBackupContainsDatabaseAttachmentsAndManifest(t *testing.T)
func TestValidateRejectsTamperedBackup(t *testing.T)
func TestRestoreCreatesPreRestoreBackup(t *testing.T)
func TestFailedRestoreLeavesCurrentWorkspaceUntouched(t *testing.T)
func TestPurgeCreatesBackupBeforePhysicalDelete(t *testing.T)
```

Run: `go test ./internal/backup ./internal/audit -v`

Expected: FAIL。

- [ ] **Step 3: 实现备份清单**

```go
type Manifest struct {
    FormatVersion int               `json:"formatVersion"`
    AppVersion    string            `json:"appVersion"`
    CreatedAt     time.Time         `json:"createdAt"`
    DatabaseSHA256 string           `json:"databaseSha256"`
    Files         map[string]string `json:"files"`
}
```

数据库使用 SQLite 在线备份能力生成一致快照；工作台备份包使用 ZIP，所有条目必须是相对路径。恢复时先解压到临时目录、验证哈希和数据库完整性，再以可回滚方式替换。

`students.Service.Purge` 必须先成功调用 `backup.Service.CreateDatabaseBackup`，再在事务中物理删除学生及关联字段值并写入审计；备份失败时不得删除。此时新增：

```text
DELETE /api/v1/students/{id}
```

- [ ] **Step 4: 实现备份恢复界面**

页面显示最近成功备份时间、备份类型、文件大小和校验状态。恢复必须依次展示：选择文件、校验结果、影响说明、输入“恢复”确认、执行结果。

- [ ] **Step 5: 验证并提交**

Run:

```powershell
go test ./internal/backup ./internal/audit -v
npm --prefix web test -- --run src/features/backup
```

Expected: PASS。

```powershell
git add internal/audit internal/backup internal/server web/src/features/backup web/src/app/routes.tsx
git commit -m "备份：实现操作审计与完整恢复"
```

---

### Task 12: 历史班级、工作台首页与数据安全提醒

**Files:**
- Create: `web/src/features/history/ArchivedClassesPage.tsx`
- Create: `web/src/features/history/ArchivedClassesPage.test.tsx`
- Create: `web/src/features/dashboard/DashboardPage.test.tsx`
- Modify: `web/src/features/dashboard/DashboardPage.tsx`
- Modify: `web/src/features/classes/ClassOverviewPage.tsx`
- Modify: `web/src/app/routes.tsx`
- Modify: `internal/classes/handler.go`

**Interfaces:**
- Consumes: 班级归档、学生统计、备份状态和操作记录 API。
- Produces: 历史班级只读查看和恢复流程。
- Produces: 首页统计与备份提醒。

- [ ] **Step 1: 写首页状态失败测试**

```tsx
it("显示当前班级、学生总数和最近备份", async () => {
  renderDashboard();
  expect(await screen.findByText("七年级一班")).toBeInTheDocument();
  expect(screen.getByText("学生 42 人")).toBeInTheDocument();
  expect(screen.getByText(/最近备份/)).toBeInTheDocument();
});

it("超过七天未备份时显示醒目提醒", async () => {
  renderDashboardWithBackupAge(8);
  expect(await screen.findByRole("alert")).toHaveTextContent("建议立即备份");
});
```

- [ ] **Step 2: 写历史班级只读失败测试**

```tsx
it("历史班级详情不显示编辑入口", async () => {
  renderArchivedClass();
  expect(await screen.findByText("只读归档")).toBeInTheDocument();
  expect(screen.queryByRole("button", { name: "编辑" })).not.toBeInTheDocument();
});
```

- [ ] **Step 3: 实现首页与历史班级页面**

首页快捷入口固定为“添加学生、导入名单、创建备份”；历史班级恢复前必须检测当前班级并提供“先归档当前班级”的明确路径。

- [ ] **Step 4: 验证并提交**

Run:

```powershell
npm --prefix web test -- --run src/features/dashboard src/features/history
npm --prefix web run build
```

Expected: PASS。

```powershell
git add web/src/features/dashboard web/src/features/history web/src/features/classes web/src/app/routes.tsx internal/classes
git commit -m "界面：完善工作台首页与历史班级"
```

---

### Task 13: 嵌入 Web 资源、Windows 构建与端到端验收

**Files:**
- Create: `internal/webassets/embed.go`
- Create: `scripts/build.ps1`
- Create: `scripts/smoke-test.ps1`
- Create: `web/playwright.config.ts`
- Create: `web/e2e/phase1.spec.ts`
- Modify: `cmd/classpilot/main.go`
- Modify: `internal/server/server.go`
- Modify: `internal/server/server_test.go`
- Modify: `web/src/app/App.tsx`
- Modify: `README.md`
- Modify: `.gitignore`

**Interfaces:**
- Consumes: 所有第一阶段 API 与 Web 页面。
- Produces: `dist/ClassPilot/ClassPilot.exe`
- Produces: 可复制运行的生产目录。

- [ ] **Step 1: 写静态资源嵌入失败测试**

```go
func TestEmbeddedIndexExists(t *testing.T) {
    body, err := fs.ReadFile(webassets.Dist, "dist/index.html")
    if err != nil { t.Fatal(err) }
    if !bytes.Contains(body, []byte("ClassPilot")) {
        t.Fatal("内嵌首页缺少产品标识")
    }
}
```

Run: `go test ./internal/webassets -v`

Expected: FAIL。

- [ ] **Step 2: 实现生产资源嵌入**

```go
//go:embed dist/*
var Dist embed.FS
```

构建脚本先执行前端测试和构建，再将 `web/dist` 同步到嵌入目录，最后执行：

```powershell
go test ./...
go build -trimpath -ldflags "-s -w -X main.version=$Version -X main.commit=$Commit" -o dist/ClassPilot/ClassPilot.exe ./cmd/classpilot
```

- [ ] **Step 3: 写 Playwright 第一阶段主流程**

```ts
test("从空工作台完成创建、导入、附件、导出、备份与归档", async ({ page }) => {
  await createCurrentClass(page);
  await importStudentsWithConflictPreview(page);
  await addStudentAndAttachment(page);
  await exportStudentList(page);
  await createWorkspaceBackup(page);
  await archiveCurrentClass(page);
  await expect(page.getByText("只读归档")).toBeVisible();
});
```

测试启动独立临时工作台目录，结束后保留失败截图和日志，不使用开发者真实数据。

- [ ] **Step 4: 实现发布包冒烟脚本**

`smoke-test.ps1` 必须：

```text
1. 在临时目录复制发布包
2. 启动 ClassPilot.exe
3. 请求健康检查并确认监听地址为 127.0.0.1
4. 检查六个业务目录和 SQLite 文件已创建
5. 停止程序
6. 复制整个目录到新位置后再次启动
7. 确认数据库完整性和健康检查通过
```

- [ ] **Step 5: 实现浏览器会话心跳与安全退出**

新增 `POST /api/v1/session/heartbeat`。前端可见时每 20 秒发送一次；所有工作台页面关闭且连续 120 秒未收到心跳时，本地服务停止接受新请求、等待在途请求完成、关闭数据库、释放工作台锁并退出。测试使用可注入时钟：

```go
func TestServerShutsDownAfterHeartbeatTimeout(t *testing.T)
func TestActiveRequestDelaysShutdown(t *testing.T)
```

- [ ] **Step 6: 执行完整验证**

Run:

```powershell
go test ./...
npm --prefix web test -- --run
npm --prefix web run build
npm --prefix web exec playwright test
powershell -ExecutionPolicy Bypass -File scripts/build.ps1
powershell -ExecutionPolicy Bypass -File scripts/smoke-test.ps1
git diff --check
git status --short
```

Expected: 全部 PASS；`git status --short` 只显示本任务预期修改。

- [ ] **Step 7: 更新中文 README**

README 必须写明：

```text
- 开发环境安装与命令
- 测试、构建和发布方法
- 工作台目录结构
- 数据备份与恢复说明
- Windows 10/11 与 Edge/Chrome 支持边界
- 真实数据不得提交 Git 的警告
```

- [ ] **Step 8: 中文提交**

```powershell
git add internal/webassets cmd/classpilot internal/server scripts web README.md .gitignore
git commit -m "发布：完成 Windows 构建与第一阶段验收"
```

---

## 2. 最终验收清单

- [ ] 空目录首次启动时自动建立工作台目录和数据库。
- [ ] 本地服务只监听 `127.0.0.1`，API 无令牌时拒绝访问。
- [ ] 创建当前班级后不能再创建第二个当前班级。
- [ ] Excel、JSON、CSV 导入均先预览、校验和处理冲突。
- [ ] 导入中任意一行失败时不留下部分数据。
- [ ] 学生、自定义字段、头像和附件可以新增、编辑、查询。
- [ ] 学生移入回收站后可恢复；物理删除前自动备份。
- [ ] 基础学生名单可选择列、横纵向打印。
- [ ] Excel、JSON、CSV 导出正确显示中文。
- [ ] 完整备份包包含 SQLite、附件和校验清单。
- [ ] 恢复失败时原工作台保持不变。
- [ ] 当前班级可归档，历史班级默认只读。
- [ ] 复制整个工作台目录后仍可启动和读取数据。
- [ ] 断网后所有核心流程仍可完成。
- [ ] Edge 与 Chrome 均通过端到端主流程。
- [ ] Git 未跟踪任何真实业务数据或构建产物。

## 3. 提交与推送节奏

每完成一个任务：

1. 运行该任务列出的自动测试。
2. 检查 `git diff --check`。
3. 检查 `git status --short`，只暂存本任务文件。
4. 使用计划中的中文提交信息提交。
5. 通过阶段审查后推送 `origin/main`。

不得将多个尚未验证的任务压成一个提交，也不得使用“更新”“修改一下”等无法追溯的提交信息。
