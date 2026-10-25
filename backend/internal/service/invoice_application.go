package service

import (
	"context"
	"fmt"
	"math"
	"net/mail"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/domain"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"

	dbent "github.com/Wei-Shaw/sub2api/ent"
)

// --- 开票申请常量与错误（定义在 domain 层，此处做别名导出） ---

const (
	InvoiceApplicationStatusPending   = domain.InvoiceApplicationStatusPending
	InvoiceApplicationStatusApproved  = domain.InvoiceApplicationStatusApproved
	InvoiceApplicationStatusIssued    = domain.InvoiceApplicationStatusIssued
	InvoiceApplicationStatusRejected  = domain.InvoiceApplicationStatusRejected
	InvoiceApplicationStatusCancelled = domain.InvoiceApplicationStatusCancelled
)

const (
	InvoiceTypeElectronicNormal  = domain.InvoiceTypeElectronicNormal
	InvoiceTypeElectronicSpecial = domain.InvoiceTypeElectronicSpecial
)

const (
	InvoiceTitleTypePersonal = domain.InvoiceTitleTypePersonal
	InvoiceTitleTypeCompany  = domain.InvoiceTitleTypeCompany
)

// MinInvoiceApplicationAmount 是一笔开票申请的最低金额（人民币元）。
const MinInvoiceApplicationAmount = domain.MinInvoiceApplicationAmount

var (
	ErrInvoiceApplicationNotFound                = domain.ErrInvoiceApplicationNotFound
	ErrInvoiceApplicationOrdersRequired          = domain.ErrInvoiceApplicationOrdersRequired
	ErrInvoiceApplicationAmountBelowMinimum      = domain.ErrInvoiceApplicationAmountBelowMinimum
	ErrInvoiceApplicationOrderNotEligible        = domain.ErrInvoiceApplicationOrderNotEligible
	ErrInvoiceApplicationOrderAlreadyClaimed     = domain.ErrInvoiceApplicationOrderAlreadyClaimed
	ErrInvoiceApplicationInvalidInvoiceType      = domain.ErrInvoiceApplicationInvalidInvoiceType
	ErrInvoiceApplicationInvalidTitleType        = domain.ErrInvoiceApplicationInvalidTitleType
	ErrInvoiceApplicationTitleRequired           = domain.ErrInvoiceApplicationTitleRequired
	ErrInvoiceApplicationTaxNumberRequired       = domain.ErrInvoiceApplicationTaxNumberRequired
	ErrInvoiceApplicationBankInfoRequired        = domain.ErrInvoiceApplicationBankInfoRequired
	ErrInvoiceApplicationEmailInvalid            = domain.ErrInvoiceApplicationEmailInvalid
	ErrInvoiceApplicationStatusTransitionInvalid = domain.ErrInvoiceApplicationStatusTransitionInvalid
	ErrInvoiceApplicationRejectReasonRequired   = domain.ErrInvoiceApplicationRejectReasonRequired
	ErrInvoiceApplicationInvoiceInfoRequired     = domain.ErrInvoiceApplicationInvoiceInfoRequired
	ErrInvoiceApplicationInvoiceAmountInvalid    = domain.ErrInvoiceApplicationInvoiceAmountInvalid
)

type InvoiceApplication = domain.InvoiceApplication
type InvoiceApplicationItem = domain.InvoiceApplicationItem

// InvoiceApplicationOrder 是一笔可开票的充值记录。
//
// PayAmount 是用户实付金额，也是开票金额的计算基数；TotalAmount 是充值入账金额
// （含赠送额度），仅用于展示。
type InvoiceApplicationOrder struct {
	OrderID     int64
	OrderNo     string
	PayAmount   float64
	TotalAmount float64
	Currency    string
	PaidAt      *time.Time
	CreatedAt   time.Time
}

// InvoiceApplicationListFilters 是开票申请列表的筛选条件。
type InvoiceApplicationListFilters struct {
	// Status 按申请状态筛选，空表示不过滤。
	Status string
	// UserID 按用户筛选，0 表示不过滤。
	UserID int64
	// Keyword 按发票抬头、纳税人识别号、用户邮箱模糊匹配。
	Keyword string
}

// InvoiceApplicationReviewUpdate 描述一次审核或开票登记需要写入的内容。
type InvoiceApplicationReviewUpdate struct {
	Status         string
	ReviewerID     int64
	RejectedReason string
	AdminRemark    string
	InvoiceNumber  string
	InvoiceCode    string
	InvoiceAmount  *float64
	IssuedAt       *time.Time
}

// CreateInvoiceApplicationInput 是用户提交开票申请的入参。
type CreateInvoiceApplicationInput struct {
	// OrderIDs 选中的充值订单 ID。
	OrderIDs []int64
	// InvoiceType 发票类型。
	InvoiceType string
	// TitleType 抬头类型。
	TitleType string
	// Title 发票抬头。
	Title string
	// TaxNumber 纳税人识别号。
	TaxNumber string
	// RegisteredAddress 注册地址。
	RegisteredAddress string
	// RegisteredPhone 注册电话。
	RegisteredPhone string
	// BankName 开户银行。
	BankName string
	// BankAccount 银行账号。
	BankAccount string
	// Email 接收发票的邮箱。
	Email string
	// Remark 备注。
	Remark string
}

// IssueInvoiceApplicationInput 是管理员登记实际开票信息的入参。
type IssueInvoiceApplicationInput struct {
	InvoiceNumber string
	InvoiceCode   string
	InvoiceAmount float64
	IssuedAt      *time.Time
	AdminRemark   string
}

// normalizedOrderIDs 去重并剔除非法 ID，同时保持用户提交的原始顺序。
func (in *CreateInvoiceApplicationInput) normalizedOrderIDs() []int64 {
	seen := make(map[int64]struct{}, len(in.OrderIDs))
	out := make([]int64, 0, len(in.OrderIDs))
	for _, id := range in.OrderIDs {
		if id <= 0 {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		out = append(out, id)
	}
	return out
}

// normalizeAndValidate 归一化并校验开票信息。
func (in *CreateInvoiceApplicationInput) normalizeAndValidate() error {
	in.InvoiceType = strings.ToLower(strings.TrimSpace(in.InvoiceType))
	if in.InvoiceType == "" {
		in.InvoiceType = InvoiceTypeElectronicNormal
	}
	switch in.InvoiceType {
	case InvoiceTypeElectronicNormal, InvoiceTypeElectronicSpecial:
	default:
		return ErrInvoiceApplicationInvalidInvoiceType
	}

	in.TitleType = strings.ToLower(strings.TrimSpace(in.TitleType))
	if in.TitleType == "" {
		in.TitleType = InvoiceTitleTypePersonal
	}
	switch in.TitleType {
	case InvoiceTitleTypePersonal, InvoiceTitleTypeCompany:
	default:
		return ErrInvoiceApplicationInvalidTitleType
	}

	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return ErrInvoiceApplicationTitleRequired
	}

	in.TaxNumber = strings.TrimSpace(in.TaxNumber)
	in.RegisteredAddress = strings.TrimSpace(in.RegisteredAddress)
	in.RegisteredPhone = strings.TrimSpace(in.RegisteredPhone)
	in.BankName = strings.TrimSpace(in.BankName)
	in.BankAccount = strings.TrimSpace(in.BankAccount)
	in.Remark = strings.TrimSpace(in.Remark)

	// 企业抬头必须有纳税人识别号；个人抬头没有纳税人识别号，忽略用户填写的内容。
	if in.TitleType == InvoiceTitleTypeCompany {
		if in.TaxNumber == "" {
			return ErrInvoiceApplicationTaxNumberRequired
		}
	} else {
		in.TaxNumber = ""
	}

	// 电子专用发票必须提供开户行信息。
	if in.InvoiceType == InvoiceTypeElectronicSpecial {
		if in.BankName == "" || in.BankAccount == "" {
			return ErrInvoiceApplicationBankInfoRequired
		}
	}

	in.Email = strings.TrimSpace(in.Email)
	if in.Email == "" || len(in.Email) > 254 {
		return ErrInvoiceApplicationEmailInvalid
	}
	addr, err := mail.ParseAddress(in.Email)
	if err != nil {
		return ErrInvoiceApplicationEmailInvalid
	}
	in.Email = addr.Address

	return nil
}

// InvoiceApplicationRepository 是开票申请的持久化接口。
type InvoiceApplicationRepository interface {
	// Create 创建申请及其明细，明细必须与申请在同一事务内写入。
	Create(ctx context.Context, app *InvoiceApplication, items []InvoiceApplicationItem) error
	// GetByID 按 ID 查询申请（含明细）。
	GetByID(ctx context.Context, id int64) (*InvoiceApplication, error)
	// ListByUser 分页查询指定用户的申请（含明细）。
	ListByUser(ctx context.Context, userID int64, params pagination.PaginationParams, filters InvoiceApplicationListFilters) ([]InvoiceApplication, *pagination.PaginationResult, error)
	// List 分页查询全部申请（含明细）。
	List(ctx context.Context, params pagination.PaginationParams, filters InvoiceApplicationListFilters) ([]InvoiceApplication, *pagination.PaginationResult, error)
	// UpdateReview 写入审核/开票结果并更新状态。
	UpdateReview(ctx context.Context, id int64, update InvoiceApplicationReviewUpdate) error
	// ReleaseItems 释放申请占用的充值记录，使其重新可申请。
	ReleaseItems(ctx context.Context, applicationID int64, releasedAt time.Time) error
	// ClaimedOrderIDs 返回仍被占用的充值订单 ID 集合。
	ClaimedOrderIDs(ctx context.Context, orderIDs []int64) (map[int64]struct{}, error)
	// ListEligiblePaymentOrders 分页查询用户可申请开票的充值记录。
	ListEligiblePaymentOrders(ctx context.Context, userID int64, params pagination.PaginationParams) ([]InvoiceApplicationOrder, *pagination.PaginationResult, error)
	// ListPaymentOrdersByIDs 按 ID 查询用户的充值记录，用于校验与计算金额。
	ListPaymentOrdersByIDs(ctx context.Context, userID int64, orderIDs []int64) ([]InvoiceApplicationOrder, error)
	// UserSnapshot 返回用户的邮箱与名称，用于冗余存储到申请单上。
	UserSnapshot(ctx context.Context, userID int64) (email string, name string, err error)
}

// InvoiceApplicationService 实现开票申请的申请、审核与开票登记。
type InvoiceApplicationService struct {
	repo       InvoiceApplicationRepository
	entClient  *dbent.Client
}

// NewInvoiceApplicationService 创建开票申请服务。
func NewInvoiceApplicationService(repo InvoiceApplicationRepository, entClient *dbent.Client) *InvoiceApplicationService {
	return &InvoiceApplicationService{
		repo:      repo,
		entClient: entClient,
	}
}

// ListEligibleOrders 列出用户可申请开票的充值记录。
func (s *InvoiceApplicationService) ListEligibleOrders(ctx context.Context, userID int64, params pagination.PaginationParams) ([]InvoiceApplicationOrder, *pagination.PaginationResult, error) {
	return s.repo.ListEligiblePaymentOrders(ctx, userID, params)
}

// CreateApplication 创建开票申请。
//
// 申请与明细在同一事务内写入，明细表 order_id 上的部分唯一索引
// （WHERE released_at IS NULL）保证一笔充值记录至多被一条未释放的申请占用，
// 并发提交时只有一条申请能成功。
func (s *InvoiceApplicationService) CreateApplication(ctx context.Context, userID int64, in CreateInvoiceApplicationInput) (*InvoiceApplication, error) {
	orderIDs := in.normalizedOrderIDs()
	if len(orderIDs) == 0 {
		return nil, ErrInvoiceApplicationOrdersRequired
	}
	if err := in.normalizeAndValidate(); err != nil {
		return nil, err
	}

	orders, err := s.repo.ListPaymentOrdersByIDs(ctx, userID, orderIDs)
	if err != nil {
		return nil, err
	}
	// 数量不一致说明有订单不存在、不属于该用户或不可开票。
	if len(orders) != len(orderIDs) {
		return nil, ErrInvoiceApplicationOrderNotEligible
	}

	byID := make(map[int64]InvoiceApplicationOrder, len(orders))
	var total float64
	for _, o := range orders {
		byID[o.OrderID] = o
		total += o.PayAmount
	}
	if total+invoiceAmountEpsilon < MinInvoiceApplicationAmount {
		return nil, ErrInvoiceApplicationAmountBelowMinimum
	}

	userEmail, userName, err := s.repo.UserSnapshot(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("load user snapshot: %w", err)
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	// 事务内再次确认未被占用，尽量在写入前给出明确错误。
	claimed, err := s.repo.ClaimedOrderIDs(txCtx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("check claimed orders: %w", err)
	}
	if len(claimed) > 0 {
		return nil, ErrInvoiceApplicationOrderAlreadyClaimed
	}

	app := &InvoiceApplication{
		UserID:            userID,
		UserEmail:         userEmail,
		UserName:          userName,
		Status:            InvoiceApplicationStatusPending,
		InvoiceType:       in.InvoiceType,
		TitleType:         in.TitleType,
		Title:             in.Title,
		TaxNumber:         in.TaxNumber,
		RegisteredAddress: in.RegisteredAddress,
		RegisteredPhone:   in.RegisteredPhone,
		BankName:          in.BankName,
		BankAccount:       in.BankAccount,
		Email:             in.Email,
		Remark:            in.Remark,
		TotalAmount:       roundInvoiceAmount(total),
	}

	items := make([]InvoiceApplicationItem, 0, len(orderIDs))
	for _, id := range orderIDs {
		o := byID[id]
		items = append(items, InvoiceApplicationItem{
			OrderID:  o.OrderID,
			OrderNo:  o.OrderNo,
			Amount:   o.PayAmount,
			Currency: o.Currency,
			PaidAt:   o.PaidAt,
		})
	}

	if err := s.repo.Create(txCtx, app, items); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit transaction: %w", err)
	}
	return app, nil
}

// ListUserApplications 分页查询用户的开票申请。
func (s *InvoiceApplicationService) ListUserApplications(ctx context.Context, userID int64, params pagination.PaginationParams, filters InvoiceApplicationListFilters) ([]InvoiceApplication, *pagination.PaginationResult, error) {
	filters.UserID = userID
	return s.repo.ListByUser(ctx, userID, params, filters)
}

// GetUserApplication 查询用户自己的开票申请详情。
func (s *InvoiceApplicationService) GetUserApplication(ctx context.Context, userID, id int64) (*InvoiceApplication, error) {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app.UserID != userID {
		return nil, ErrInvoiceApplicationNotFound
	}
	return app, nil
}

// CancelApplication 用户取消自己的待审核申请，并释放占用的充值记录。
func (s *InvoiceApplicationService) CancelApplication(ctx context.Context, userID, id int64) error {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if app.UserID != userID {
		return ErrInvoiceApplicationNotFound
	}
	if app.Status != InvoiceApplicationStatusPending {
		return ErrInvoiceApplicationStatusTransitionInvalid
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	if err := s.repo.UpdateReview(txCtx, id, InvoiceApplicationReviewUpdate{
		Status: InvoiceApplicationStatusCancelled,
	}); err != nil {
		return err
	}
	if err := s.repo.ReleaseItems(txCtx, id, time.Now()); err != nil {
		return fmt.Errorf("release claimed orders: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// AdminList 分页查询全部开票申请。
func (s *InvoiceApplicationService) AdminList(ctx context.Context, params pagination.PaginationParams, filters InvoiceApplicationListFilters) ([]InvoiceApplication, *pagination.PaginationResult, error) {
	return s.repo.List(ctx, params, filters)
}

// AdminGet 查询开票申请详情。
func (s *InvoiceApplicationService) AdminGet(ctx context.Context, id int64) (*InvoiceApplication, error) {
	return s.repo.GetByID(ctx, id)
}

// AdminApprove 审核通过一条待审核申请。
func (s *InvoiceApplicationService) AdminApprove(ctx context.Context, id, adminID int64) error {
	return s.adminTransition(ctx, id, InvoiceApplicationStatusApproved, InvoiceApplicationReviewUpdate{
		ReviewerID: adminID,
	})
}

// AdminReject 驳回一条待审核申请，并释放占用的充值记录。
func (s *InvoiceApplicationService) AdminReject(ctx context.Context, id, adminID int64, reason, remark string) error {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return ErrInvoiceApplicationRejectReasonRequired
	}
	return s.adminTransition(ctx, id, InvoiceApplicationStatusRejected, InvoiceApplicationReviewUpdate{
		ReviewerID:     adminID,
		RejectedReason: reason,
		AdminRemark:    strings.TrimSpace(remark),
	})
}

// AdminIssue 在审核通过的申请上登记实际开具的发票信息。
func (s *InvoiceApplicationService) AdminIssue(ctx context.Context, id, adminID int64, in IssueInvoiceApplicationInput) error {
	invoiceNumber := strings.TrimSpace(in.InvoiceNumber)
	if invoiceNumber == "" {
		return ErrInvoiceApplicationInvoiceInfoRequired
	}
	if in.InvoiceAmount <= 0 {
		return ErrInvoiceApplicationInvoiceAmountInvalid
	}

	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if app.Status != InvoiceApplicationStatusApproved {
		return ErrInvoiceApplicationStatusTransitionInvalid
	}

	amount := roundInvoiceAmount(in.InvoiceAmount)
	issuedAt := in.IssuedAt
	if issuedAt == nil {
		now := time.Now()
		issuedAt = &now
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	if err := s.repo.UpdateReview(txCtx, id, InvoiceApplicationReviewUpdate{
		Status:        InvoiceApplicationStatusIssued,
		ReviewerID:    adminID,
		InvoiceNumber: invoiceNumber,
		InvoiceCode:   strings.TrimSpace(in.InvoiceCode),
		InvoiceAmount: &amount,
		IssuedAt:      issuedAt,
		AdminRemark:   strings.TrimSpace(in.AdminRemark),
	}); err != nil {
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// adminTransition 执行一次状态流转，仅允许从待审核状态流转。
func (s *InvoiceApplicationService) adminTransition(ctx context.Context, id int64, target string, update InvoiceApplicationReviewUpdate) error {
	app, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if app.Status != InvoiceApplicationStatusPending {
		return ErrInvoiceApplicationStatusTransitionInvalid
	}

	tx, err := s.entClient.Tx(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()
	txCtx := dbent.NewTxContext(ctx, tx)

	update.Status = target
	if err := s.repo.UpdateReview(txCtx, id, update); err != nil {
		return err
	}
	// 驳回后释放占用，用户可以基于同一笔充值重新申请。
	if target == InvoiceApplicationStatusRejected {
		if err := s.repo.ReleaseItems(txCtx, id, time.Now()); err != nil {
			return fmt.Errorf("release claimed orders: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

// invoiceAmountEpsilon 是金额比较时允许的浮点误差。
const invoiceAmountEpsilon = 1e-6

// roundInvoiceAmount 把金额规整到分（两位小数）。
func roundInvoiceAmount(v float64) float64 {
	return math.Round(v*100) / 100
}
