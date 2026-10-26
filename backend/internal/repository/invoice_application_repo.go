package repository

import (
	"context"
	"fmt"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/invoiceapplication"
	"github.com/Wei-Shaw/sub2api/ent/invoiceapplicationitem"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/ent/user"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/service"

	entsql "entgo.io/ent/dialect/sql"
)

type invoiceApplicationRepository struct {
	client *dbent.Client
}

// NewInvoiceApplicationRepository 创建开票申请仓储。
func NewInvoiceApplicationRepository(client *dbent.Client) service.InvoiceApplicationRepository {
	return &invoiceApplicationRepository{client: client}
}

// invoiceApplicationCurrencyCNYPredicate 判断订单的支付币种是否为人民币。
//
// 该表达式与 service.PaymentOrderCurrency 的语义保持一致：
//   - provider_snapshot 为空或未声明币种时取默认币种 CNY；
//   - 声明的币种必须是 3 位字母，否则同样回落到默认币种 CNY；
//   - 只有显式声明为非人民币币种（如 USD）的订单才不可开票。
const invoiceApplicationCurrencyCNYPredicate = `COALESCE(
    CASE WHEN UPPER(provider_snapshot->>'currency') ~ '^[A-Z]{3}$'
         THEN UPPER(provider_snapshot->>'currency')
    END,
    'CNY'
) = 'CNY'`

// eligiblePaymentOrderQuery 构造「用户可申请开票的充值记录」查询。
//
// 条件：属于该用户、余额充值类型、已完成（已收款且已入账）、人民币计价。
func (r *invoiceApplicationRepository) eligiblePaymentOrderQuery(userID int64) *dbent.PaymentOrderQuery {
	return r.client.PaymentOrder.Query().Where(
		paymentorder.UserIDEQ(userID),
		paymentorder.OrderTypeEQ(payment.OrderTypeBalance),
		paymentorder.StatusEQ(service.OrderStatusCompleted),
		// 原始 SQL 谓词需要包一层 selector 修改器才能塞进 ent 的 Where。
		func(s *entsql.Selector) {
			s.Where(entsql.ExprP(invoiceApplicationCurrencyCNYPredicate))
		},
	)
}

// ListEligiblePaymentOrders 分页查询用户可申请开票的充值记录，并排除已被占用的订单。
func (r *invoiceApplicationRepository) ListEligiblePaymentOrders(ctx context.Context, userID int64, params pagination.PaginationParams) ([]service.InvoiceApplicationOrder, *pagination.PaginationResult, error) {
	claimed, err := r.claimedOrderIDsForUser(ctx, userID)
	if err != nil {
		return nil, nil, fmt.Errorf("list claimed order ids: %w", err)
	}

	q := r.eligiblePaymentOrderQuery(userID)
	if len(claimed) > 0 {
		q = q.Where(paymentorder.IDNotIn(claimed...))
	}

	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count eligible payment orders: %w", err)
	}

	orders, err := q.
		Order(dbent.Desc(paymentorder.FieldPaidAt), dbent.Desc(paymentorder.FieldID)).
		Limit(params.Limit()).
		Offset(params.Offset()).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("query eligible payment orders: %w", err)
	}

	result := &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: params.Limit(),
	}
	if result.PageSize > 0 {
		result.Pages = int((int64(total) + int64(result.PageSize) - 1) / int64(result.PageSize))
	}

	out := make([]service.InvoiceApplicationOrder, 0, len(orders))
	for _, o := range orders {
		out = append(out, paymentOrderToInvoiceOrder(o))
	}
	return out, result, nil
}

// ListPaymentOrdersByIDs 按 ID 查询用户的充值记录。
func (r *invoiceApplicationRepository) ListPaymentOrdersByIDs(ctx context.Context, userID int64, orderIDs []int64) ([]service.InvoiceApplicationOrder, error) {
	if len(orderIDs) == 0 {
		return nil, nil
	}
	orders, err := r.client.PaymentOrder.Query().
		Where(
			paymentorder.UserIDEQ(userID),
			paymentorder.IDIn(orderIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query payment orders by ids: %w", err)
	}
	out := make([]service.InvoiceApplicationOrder, 0, len(orders))
	for _, o := range orders {
		out = append(out, paymentOrderToInvoiceOrder(o))
	}
	return out, nil
}

// paymentOrderToInvoiceOrder 把充值订单转换为开票视角的记录。
func paymentOrderToInvoiceOrder(o *dbent.PaymentOrder) service.InvoiceApplicationOrder {
	return service.InvoiceApplicationOrder{
		OrderID:     o.ID,
		OrderNo:     o.OutTradeNo,
		PayAmount:   o.PayAmount,
		TotalAmount: o.Amount,
		Currency:    service.PaymentOrderCurrency(o),
		PaidAt:      o.PaidAt,
		CreatedAt:   o.CreatedAt,
	}
}

// claimedOrderIDsForUser 返回该用户仍被开票申请占用的充值订单 ID。
func (r *invoiceApplicationRepository) claimedOrderIDsForUser(ctx context.Context, userID int64) ([]int64, error) {
	items, err := r.client.InvoiceApplication.Query().
		Where(invoiceapplication.UserIDEQ(userID)).
		QueryItems().
		Where(invoiceapplicationitem.ReleasedAtIsNil()).
		All(ctx)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(items))
	for _, it := range items {
		ids = append(ids, it.OrderID)
	}
	return ids, nil
}

// ClaimedOrderIDs 返回仍被占用的充值订单 ID 集合。
func (r *invoiceApplicationRepository) ClaimedOrderIDs(ctx context.Context, orderIDs []int64) (map[int64]struct{}, error) {
	if len(orderIDs) == 0 {
		return map[int64]struct{}{}, nil
	}
	items, err := r.client.InvoiceApplicationItem.Query().
		Where(
			invoiceapplicationitem.ReleasedAtIsNil(),
			invoiceapplicationitem.OrderIDIn(orderIDs...),
		).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("query claimed order ids: %w", err)
	}
	out := make(map[int64]struct{}, len(items))
	for _, it := range items {
		out[it.OrderID] = struct{}{}
	}
	return out, nil
}

// Create 创建申请及其明细。
//
// 明细写入依赖 invoice_application_items 上的部分唯一索引
// （order_id WHERE released_at IS NULL）：并发提交同一笔充值时，
// 只有一条申请能成功，另一条会因唯一约束冲突被拒绝。
func (r *invoiceApplicationRepository) Create(ctx context.Context, app *service.InvoiceApplication, items []service.InvoiceApplicationItem) error {
	client := clientFromContext(ctx, r.client)

	builder := client.InvoiceApplication.Create().
		SetUserID(app.UserID).
		SetUserEmail(app.UserEmail).
		SetUserName(app.UserName).
		SetStatus(app.Status).
		SetInvoiceType(app.InvoiceType).
		SetTitleType(app.TitleType).
		SetTitle(app.Title).
		SetTaxNumber(app.TaxNumber).
		SetRegisteredAddress(app.RegisteredAddress).
		SetRegisteredPhone(app.RegisteredPhone).
		SetBankName(app.BankName).
		SetBankAccount(app.BankAccount).
		SetEmail(app.Email).
		SetRemark(app.Remark).
		SetTotalAmount(app.TotalAmount)

	created, err := builder.Save(ctx)
	if err != nil {
		return fmt.Errorf("create invoice application: %w", err)
	}

	if len(items) > 0 {
		itemBuilders := make([]*dbent.InvoiceApplicationItemCreate, 0, len(items))
		for _, it := range items {
			ib := client.InvoiceApplicationItem.Create().
				SetApplicationID(created.ID).
				SetOrderID(it.OrderID).
				SetOrderNo(it.OrderNo).
				SetAmount(it.Amount).
				SetCurrency(it.Currency)
			if it.PaidAt != nil {
				ib = ib.SetPaidAt(*it.PaidAt)
			}
			itemBuilders = append(itemBuilders, ib)
		}
		if _, err := client.InvoiceApplicationItem.CreateBulk(itemBuilders...).Save(ctx); err != nil {
			// 唯一约束冲突：同一笔充值记录已被另一条申请占用。
			return translatePersistenceError(err, nil, service.ErrInvoiceApplicationOrderAlreadyClaimed)
		}
	}

	applyInvoiceApplicationEntityToService(app, created)
	app.Items = make([]service.InvoiceApplicationItem, 0, len(items))
	app.Items = append(app.Items, items...)
	return nil
}

// GetByID 按 ID 查询申请（含明细）。
func (r *invoiceApplicationRepository) GetByID(ctx context.Context, id int64) (*service.InvoiceApplication, error) {
	m, err := r.client.InvoiceApplication.Query().
		Where(invoiceapplication.IDEQ(id)).
		WithItems().
		Only(ctx)
	if err != nil {
		return nil, translatePersistenceError(err, service.ErrInvoiceApplicationNotFound, nil)
	}
	return invoiceApplicationEntityToService(m), nil
}

// ListByUser 分页查询指定用户的申请（含明细）。
func (r *invoiceApplicationRepository) ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams, filters service.InvoiceApplicationListFilters) ([]service.InvoiceApplication, *pagination.PaginationResult, error) {
	q := r.client.InvoiceApplication.Query().Where(invoiceapplication.UserIDEQ(userID))
	q = applyInvoiceApplicationFilters(q, filters)
	return r.listWithItems(ctx, q, params)
}

// List 分页查询全部申请（含明细）。
func (r *invoiceApplicationRepository) List(ctx context.Context, params pagination.PaginationParams, filters service.InvoiceApplicationListFilters) ([]service.InvoiceApplication, *pagination.PaginationResult, error) {
	q := r.client.InvoiceApplication.Query()
	q = applyInvoiceApplicationFilters(q, filters)
	return r.listWithItems(ctx, q, params)
}

// applyInvoiceApplicationFilters 应用列表筛选条件。
func applyInvoiceApplicationFilters(q *dbent.InvoiceApplicationQuery, filters service.InvoiceApplicationListFilters) *dbent.InvoiceApplicationQuery {
	if filters.Status != "" {
		q = q.Where(invoiceapplication.StatusEQ(filters.Status))
	}
	if filters.UserID > 0 {
		q = q.Where(invoiceapplication.UserIDEQ(filters.UserID))
	}
	if filters.Keyword != "" {
		q = q.Where(invoiceapplication.Or(
			invoiceapplication.TitleContainsFold(filters.Keyword),
			invoiceapplication.TaxNumberContainsFold(filters.Keyword),
			invoiceapplication.UserEmailContainsFold(filters.Keyword),
			invoiceapplication.UserNameContainsFold(filters.Keyword),
		))
	}
	return q
}

// listWithItems 执行分页查询并附带明细。
func (r *invoiceApplicationRepository) listWithItems(ctx context.Context, q *dbent.InvoiceApplicationQuery, params pagination.PaginationParams) ([]service.InvoiceApplication, *pagination.PaginationResult, error) {
	total, err := q.Clone().Count(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("count invoice applications: %w", err)
	}

	apps, err := q.
		WithItems(func(iq *dbent.InvoiceApplicationItemQuery) {
			iq.Order(dbent.Asc(invoiceapplicationitem.FieldID))
		}).
		Order(dbent.Desc(invoiceapplication.FieldCreatedAt), dbent.Desc(invoiceapplication.FieldID)).
		Limit(params.Limit()).
		Offset(params.Offset()).
		All(ctx)
	if err != nil {
		return nil, nil, fmt.Errorf("query invoice applications: %w", err)
	}

	result := &pagination.PaginationResult{
		Total:    int64(total),
		Page:     params.Page,
		PageSize: params.Limit(),
	}
	if result.PageSize > 0 {
		result.Pages = int((int64(total) + int64(result.PageSize) - 1) / int64(result.PageSize))
	}

	out := make([]service.InvoiceApplication, 0, len(apps))
	for _, a := range apps {
		out = append(out, *invoiceApplicationEntityToService(a))
	}
	return out, result, nil
}

// UpdateReview 写入审核/开票结果并更新状态。
func (r *invoiceApplicationRepository) UpdateReview(ctx context.Context, id int64, update service.InvoiceApplicationReviewUpdate) error {
	client := clientFromContext(ctx, r.client)

	builder := client.InvoiceApplication.UpdateOneID(id).
		SetStatus(update.Status)

	if update.ReviewerID > 0 {
		builder = builder.SetReviewedBy(update.ReviewerID)
	}
	now := time.Now()
	builder = builder.SetReviewedAt(now)
	if update.RejectedReason != "" {
		builder = builder.SetRejectedReason(update.RejectedReason)
	}
	if update.AdminRemark != "" {
		builder = builder.SetAdminRemark(update.AdminRemark)
	}
	if update.InvoiceNumber != "" {
		builder = builder.SetInvoiceNumber(update.InvoiceNumber)
	}
	if update.InvoiceCode != "" {
		builder = builder.SetInvoiceCode(update.InvoiceCode)
	}
	if update.InvoiceAmount != nil {
		builder = builder.SetInvoiceAmount(*update.InvoiceAmount)
	}
	if update.IssuedAt != nil {
		builder = builder.SetIssuedAt(*update.IssuedAt)
	}

	if _, err := builder.Save(ctx); err != nil {
		return translatePersistenceError(err, service.ErrInvoiceApplicationNotFound, nil)
	}
	return nil
}

// ReleaseItems 释放申请占用的充值记录。
func (r *invoiceApplicationRepository) ReleaseItems(ctx context.Context, applicationID int64, releasedAt time.Time) error {
	client := clientFromContext(ctx, r.client)
	_, err := client.InvoiceApplicationItem.Update().
		Where(
			invoiceapplicationitem.ApplicationIDEQ(applicationID),
			invoiceapplicationitem.ReleasedAtIsNil(),
		).
		SetReleasedAt(releasedAt).
		Save(ctx)
	if err != nil {
		return fmt.Errorf("release invoice application items: %w", err)
	}
	return nil
}

// UserSnapshot 返回用户的邮箱与名称。
func (r *invoiceApplicationRepository) UserSnapshot(ctx context.Context, userID int64) (string, string, error) {
	u, err := r.client.User.Query().
		Where(user.IDEQ(userID)).
		Only(ctx)
	if err != nil {
		return "", "", translatePersistenceError(err, nil, nil)
	}
	return u.Email, u.Username, nil
}

// invoiceApplicationEntityToService 把 ent 实体转换为 service 结构体。
func invoiceApplicationEntityToService(m *dbent.InvoiceApplication) *service.InvoiceApplication {
	if m == nil {
		return nil
	}
	out := &service.InvoiceApplication{
		ID:                m.ID,
		UserID:            m.UserID,
		UserEmail:         m.UserEmail,
		UserName:          m.UserName,
		Status:            m.Status,
		InvoiceType:       m.InvoiceType,
		TitleType:         m.TitleType,
		Title:             m.Title,
		TaxNumber:         m.TaxNumber,
		RegisteredAddress: m.RegisteredAddress,
		RegisteredPhone:   m.RegisteredPhone,
		BankName:          m.BankName,
		BankAccount:       m.BankAccount,
		Email:             m.Email,
		Remark:            m.Remark,
		TotalAmount:       m.TotalAmount,
		InvoiceNumber:     m.InvoiceNumber,
		InvoiceCode:       m.InvoiceCode,
		RejectedReason:    m.RejectedReason,
		AdminRemark:       m.AdminRemark,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
	if m.InvoiceAmount != nil {
		amount := *m.InvoiceAmount
		out.InvoiceAmount = &amount
	}
	if m.IssuedAt != nil {
		issuedAt := *m.IssuedAt
		out.IssuedAt = &issuedAt
	}
	if m.ReviewedBy != nil {
		reviewedBy := *m.ReviewedBy
		out.ReviewedBy = &reviewedBy
	}
	if m.ReviewedAt != nil {
		reviewedAt := *m.ReviewedAt
		out.ReviewedAt = &reviewedAt
	}
	if m.Edges.Items != nil {
		out.Items = make([]service.InvoiceApplicationItem, 0, len(m.Edges.Items))
		for _, it := range m.Edges.Items {
			item := service.InvoiceApplicationItem{
				ID:            it.ID,
				ApplicationID: it.ApplicationID,
				OrderID:       it.OrderID,
				OrderNo:       it.OrderNo,
				Amount:        it.Amount,
				Currency:      it.Currency,
				CreatedAt:     it.CreatedAt,
			}
			if it.PaidAt != nil {
				paidAt := *it.PaidAt
				item.PaidAt = &paidAt
			}
			if it.ReleasedAt != nil {
				releasedAt := *it.ReleasedAt
				item.ReleasedAt = &releasedAt
			}
			out.Items = append(out.Items, item)
		}
	}
	return out
}

// applyInvoiceApplicationEntityToService 用新建的实体回填申请单。
func applyInvoiceApplicationEntityToService(app *service.InvoiceApplication, m *dbent.InvoiceApplication) {
	if app == nil || m == nil {
		return
	}
	app.ID = m.ID
	app.Status = m.Status
	app.CreatedAt = m.CreatedAt
	app.UpdatedAt = m.UpdatedAt
}
