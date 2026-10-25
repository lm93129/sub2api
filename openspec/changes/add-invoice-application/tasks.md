## 1. 数据模型与持久化

- [x] 1.1 在 `backend/ent/schema/invoice_application.go` 新增 `InvoiceApplication`（申请主表）与 `InvoiceApplicationItem`（申请明细）两个实体：表名 `invoice_applications` / `invoice_application_items`，金额 `decimal(20,2)`，时间 `timestamptz`，时间戳用 `mixins.TimeMixin{}`
- [x] 1.2 关联策略：不建指向 `users` / `payment_orders` 的外键。`payment_orders` 是硬删除，建外键会阻断订单删除并改变既有行为；用户被删除后开票凭证仍需留存审计。因此用户邮箱/用户名、订单号/实付金额/支付时间一律以快照冗余存储
- [x] 1.3 明细表 `order_id` 建部分唯一索引（`WHERE released_at IS NULL`），把「一笔充值记录至多被一条未释放申请占用」的约束落在数据库层。释放语义：`released_at IS NULL` 表示仍被占用；驳回或取消时写入该时间释放占用，明细行保留供审计
- [x] 1.4 执行 `go generate ./ent`，生成 22 个文件（含 `invoiceapplication*.go`、两个子包、`client.go`/`mutation.go`/`migrate/schema.go` 等 9 个共享文件改动），已提交
- [x] 1.5 修正 ent 生成的两处陷阱：同字段声明两个索引会被拆成「非唯一带 Where」+「唯一不带 Where」而丢失部分唯一语义（删掉重复的 `index.Fields("order_id")`）；`application_id` 字段与 edge 并存会生成重复的 `SetApplicationID` 方法导致生成代码无法编译（改用 `.Field("application_id")`）
- [x] 1.6 新增迁移 `backend/migrations/248_invoice_applications.sql`：建两张表 + 6 个普通索引 + `idx_invoice_application_items_order_unique_active` 部分唯一索引；全部语句幂等（`IF NOT EXISTS`）

## 2. 领域常量与业务规则

- [x] 2.1 新增 `backend/internal/domain/invoice_application.go`：状态常量（`PENDING`/`APPROVED`/`ISSUED`/`REJECTED`/`CANCELLED`）、发票类型（`electron_normal`/`electron_special`）、抬头类型（`personal`/`company`）、最低金额常量 `MinInvoiceApplicationAmount = 300.0`
- [x] 2.2 定义业务错误码：`INVOICE_APPLICATION_NOT_FOUND`、`_ORDERS_REQUIRED`、`_AMOUNT_BELOW_MINIMUM`、`_ORDER_NOT_ELIGIBLE`、`_ORDER_ALREADY_CLAIMED`、`_INVOICE_TYPE_INVALID`、`_TITLE_TYPE_INVALID`、`_TITLE_REQUIRED`、`_TAX_NUMBER_REQUIRED`、`_BANK_INFO_REQUIRED`、`_EMAIL_INVALID`、`_STATUS_TRANSITION_INVALID`、`_REJECT_REASON_REQUIRED`、`_INVOICE_INFO_REQUIRED`、`_INVOICE_AMOUNT_INVALID`
- [x] 2.3 定义 `InvoiceApplication` / `InvoiceApplicationItem` 结构体与 `IsTerminal()` 终态判定

## 3. 持久化层

- [x] 3.1 新增 `backend/internal/repository/invoice_application_repo.go`：实现 `Create`（申请与明细同事务写入）、`GetByID`（含明细）、`ListByUser` / `List`（含明细、状态/用户/关键字筛选）、`UpdateReview`、`ReleaseItems`、`ClaimedOrderIDs`、`ListEligiblePaymentOrders`、`ListPaymentOrdersByIDs`、`UserSnapshot`
- [x] 3.2 可开票订单筛选条件：属于该用户、`order_type='balance'`、状态 `COMPLETED`（已收款且已入账）、人民币计价、且未被未释放申请占用
- [x] 3.3 币种判定用原始 SQL 谓词 `COALESCE(CASE WHEN UPPER(provider_snapshot->>'currency') ~ '^[A-Z]{3}$' THEN UPPER(provider_snapshot->>'currency') END, 'CNY') = 'CNY'`，与 `service.PaymentOrderCurrency` 的归一化语义完全一致（快照缺失/空/非法币种均回落到默认币种 CNY）
- [x] 3.4 唯一约束冲突经 `translatePersistenceError` 映射为 `ErrInvoiceApplicationOrderAlreadyClaimed`，不向业务层泄露数据库细节

## 4. 服务层

- [x] 4.1 新增 `backend/internal/service/invoice_application.go`：定义 `InvoiceApplicationRepository` 接口与 `InvoiceApplicationService`
- [x] 4.2 `CreateApplication`：去重订单 ID（保持提交顺序）→ 校验订单归属与可开票性 → 按 `pay_amount` 合计校验 ≥ 300 元 → 事务内二次确认未占用 → 同事务写申请与明细 → 唯一索引兜底并发
- [x] 4.3 开票信息归一化与校验：企业抬头必须有纳税人识别号（个人抬头强制清空）、电子专用发票必须提供开户银行与银行账号、接收邮箱走 `mail.ParseAddress` 校验
- [x] 4.4 状态机：`adminTransition` 仅允许从 `PENDING` 流转；`AdminReject` 必须填写原因且驳回时释放占用；`AdminIssue` 仅允许对 `APPROVED` 申请登记开票信息，实际开票金额可与申请金额不同
- [x] 4.5 `CancelApplication` 仅允许用户取消自己的 `PENDING` 申请并释放占用

## 5. 接口层

- [x] 5.1 新增 `backend/internal/handler/dto/invoice_application.go`：`InvoiceApplication`、`InvoiceApplicationItem`、`InvoiceApplicationOrder` 三个 DTO 与转换函数
- [x] 5.2 新增 `backend/internal/handler/invoice_application_handler.go`（用户端）：可开票订单列表、我的申请列表、详情、提交申请、取消申请
- [x] 5.3 新增 `backend/internal/handler/admin/invoice_application_handler.go`（管理端）：列表筛选、详情、审核通过、驳回、登记开票信息

## 6. 路由、依赖注入与管理端权限

- [x] 6.1 `backend/internal/server/routes/user.go` 注册 `/api/v1/invoice/*`：`GET /eligible-orders`、`GET /applications`、`POST /applications`、`GET /applications/:id`、`POST /applications/:id/cancel`
- [x] 6.2 `backend/internal/server/routes/admin.go` 新增 `registerInvoiceApplicationRoutes`，注册 `/api/v1/admin/invoice/applications/*` 的列表、详情、approve、reject、issue
- [x] 6.3 `backend/internal/handler/handler.go` 的 `Handlers` 与 `AdminHandlers` 结构体各增加一个字段
- [x] 6.4 `backend/internal/handler/wire.go`、`backend/internal/service/wire.go`、`backend/internal/repository/wire.go` 三个 provider set 注册新组件
- [x] 6.5 `backend/internal/pkg/adminaccess/catalog.go` 注册 `invoice.read` / `invoice.review` / `invoice.issue` 三个权限点，避免受限管理密钥被拒之门外

## 7. 前端

- [ ] 7.1 `frontend/src/api/invoice.ts` 用户端与管理端 API 模块，`frontend/src/types/index.ts` 新增类型
- [ ] 7.2 `frontend/src/views/user/InvoiceApplicationsView.vue`：可开票记录勾选、合计实时显示、未达 300 元阻止提交、开票信息表单（企业必填税号、专票必填开户行）、我的申请列表与取消
- [ ] 7.3 `frontend/src/views/admin/InvoiceApplicationsView.vue`：状态与关键字筛选、分页、详情对话框（含充值明细）、通过/驳回/登记开票操作
- [ ] 7.4 用户路由与管理端路由各增一条懒加载路由
- [ ] 7.5 `frontend/src/components/layout/AppSidebar.vue` 管理端与用户端各增一个菜单入口
- [ ] 7.6 所有 locale 文件补齐 i18n key（`localeKeyCompleteness.spec.ts` 会校验缺失）
- [ ] 7.7 管理端列表页 vitest 用例

## 8. 测试与验证

- [x] 8.1 后端单元测试：新增 `backend/tests/invoiceapplication/invoice_application_test.go`（外置测试包 `invoiceapplication_test`），覆盖金额下限（299.99 被拒 / 300 通过）、企业必填税号、专票必填开户行、个人抬头清空税号、非法发票/抬头类型、空抬头、非法邮箱、订单 ID 去重、不可开票订单、已被占用订单被拒、取消释放占用、状态机非法流转被拒、驳回必须填原因且释放占用、登记开票要求已审核且金额可为正、开票金额与申请金额不同
  - 说明：测试放在独立包而非 `internal/service` 目录下。该目录有 868 个测试文件，其「含测试文件的编译变体」在 4C/4G 机器上会超出 cgroup 内存上限而 OOM（已实测：移除本次测试文件后基线同样 OOM）。独立包只依赖已编译缓存的 service 包，可稳定运行。
- [x] 8.2 端到端验证：本地空库执行全部 248 个迁移 → 两张表与 5 个索引正确建立 → 验证部分唯一索引拒绝同一 `order_id` 重复占用 → 释放占用后可再次占用
- [x] 8.3 冷构建校验：`cd backend && go clean -cache && make build` → exit 0，耗时 7m30s，产物 `backend/bin/server` 114MB 可执行（`./bin/server --version` 输出 `Sub2API 0.2.18`）
- [ ] 8.4 `cd backend && make test` —— 注意：本机内存不足以编译 `internal/service` 的测试变体（基线即 OOM），完整测试套件需在更大内存的机器上执行。本次新增的两组测试已单独跑通：
  - `go test -tags=unit ./tests/invoiceapplication/` → ok（17 个用例）
  - `go test ./internal/pkg/adminaccess/...` → ok（管理端权限目录登记与权限独立性）
- [x] 8.5 `cd frontend && pnpm typecheck && pnpm lint:check && pnpm check:i18n` → 三项全绿
  - 注意：本机 Node 默认堆上限约 1GB，`vue-tsc` 会 OOM，需 `NODE_OPTIONS=--max-old-space-size=3072`
