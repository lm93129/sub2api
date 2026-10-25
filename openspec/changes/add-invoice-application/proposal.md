## Why

sub2api 的余额充值走 `payment_orders` 表，但用户充值完成后没有任何途径申请发票：平台在财务侧无法闭环，用户只能通过客服人工核对订单、手工登记，抬头信息要靠邮件往返收集，且客服无法判断哪些充值已经被开过票。同一笔充值可以被反复提交开票申请，造成重复开票与对账风险。

## What Changes

- 新增「开票申请」能力：用户在个人中心查看**可开票的充值记录**（本人名下、已支付成功的余额充值订单、人民币计价），勾选若干笔后提交申请；所选记录实付金额合计 **≥ 300 元**才允许提交。
- 每条申请采集中国增值税发票常见信息：发票类型（电子普通发票 / 电子专用发票）、抬头类型（企业 / 个人）、发票抬头、纳税人识别号、注册地址、注册电话、开户银行、银行账号、接收发票邮箱、备注。
- **一笔充值记录至多被一条开票申请占用**，该约束由数据库层的部分唯一索引保证，而非仅靠 Go 代码判断。申请被驳回或由用户取消时释放占用并保留记录以供审计，用户可基于同一笔充值重新申请。
- 新增申请状态机：`PENDING`（待审核）→ `APPROVED`（已审核待开票）→ `ISSUED`（已开票）；`PENDING` → `REJECTED`（已驳回）；`PENDING` → `CANCELLED`（用户已取消）。终态申请不可再流转。
- 管理端新增「发票申请」菜单与列表页：按状态、用户、关键字筛选，查看申请内容与关联的充值明细，审核通过 / 驳回（必须填写原因），并登记发票号码、发票代码、开票金额与开票日期。
- 用户端新增「发票申请」页面：勾选可开票充值记录并提交申请，查看自己的申请列表与状态。
- 开票金额下限 300 元为可覆盖的常量，默认 300；实际开票金额允许管理员在登记时与申请金额不同。
- 充值、支付、退款、余额入账等既有流程**不做任何改动**。

### Non-goals

- 不对接税务平台或第三方电子发票服务，本变更只管理申请单，人工开票后在后台登记发票信息。
- 不实现红冲、作废、退票流程。
- 不支持跨用户合并开票，也不支持对一笔充值订单做部分金额开票。
- 不改变现有支付渠道与结算流程。

## Capabilities

### New Capabilities

- `invoice-application`：用户端开票申请（可开票充值记录、勾选提交、金额下限、抬头信息、申请状态与取消）与管理端审核登记（列表筛选、审核通过/驳回、登记发票信息）。

### Modified Capabilities

<!-- openspec/specs 目前为空，没有既有能力需要修改。 -->

## Impact

- **数据库**：新增 `invoice_applications`、`invoice_application_items` 两张表与迁移 `backend/migrations/248_invoice_applications.sql`；`invoice_application_items.order_id` 建部分唯一索引（`WHERE released_at IS NULL`）保证一笔充值记录只被占用一次。
- **后端**：新增 ent schema `backend/ent/schema/invoice_application.go`（含 `invoice_application_item`），执行 `go generate ./ent` 并提交生成代码；新增 `internal/repository/invoice_application_repo.go`、`internal/service/invoice_application.go`、`internal/handler/invoice_application.go`（用户端）与 `internal/handler/admin/invoice_application.go`（管理端）；`internal/server/routes/` 注册用户端与管理端路由；`internal/handler/wire.go`、`cmd/server/wire.go` 注入新组件。
- **管理端 API**：新增 `/api/v1/invoice/*`（用户端）与 `/api/v1/admin/invoice/applications/*`（管理端）；管理端路由必须登记到 `backend/internal/pkg/adminaccess/catalog.go`，否则受限管理密钥无法访问且访问测试失败。
- **前端**：新增 `frontend/src/api/invoice.ts`、`frontend/src/views/user/InvoiceApplicationsView.vue`、`frontend/src/views/admin/InvoiceApplicationsView.vue`；`frontend/src/types/index.ts` 新增类型；用户路由与管理端路由各增一条；管理端侧边栏新增「发票申请」菜单项；所有 locale 文件补齐 i18n key。
- **上游影响**：无。开票申请只读 `payment_orders`，不改变支付链路。
