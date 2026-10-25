package domain

import (
	"time"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
)

// 开票申请状态。
//
// PENDING → APPROVED → ISSUED 为正常链路；PENDING → REJECTED（管理员驳回）与
// PENDING → CANCELLED（用户取消）为终止分支。APPROVED / ISSUED / REJECTED /
// CANCELLED 均为终态，不可再流转。
const (
	InvoiceApplicationStatusPending   = "PENDING"
	InvoiceApplicationStatusApproved  = "APPROVED"
	InvoiceApplicationStatusIssued    = "ISSUED"
	InvoiceApplicationStatusRejected  = "REJECTED"
	InvoiceApplicationStatusCancelled = "CANCELLED"
)

// 发票类型。
const (
	InvoiceTypeElectronicNormal  = "electron_normal"
	InvoiceTypeElectronicSpecial = "electron_special"
)

// 发票抬头类型。
const (
	InvoiceTitleTypePersonal = "personal"
	InvoiceTitleTypeCompany  = "company"
)

// MinInvoiceApplicationAmount 是一笔开票申请允许的最低金额（人民币元）。
//
// 低于该金额的订单不单独开票，用户需累积勾选到该金额以上才能提交申请。
const MinInvoiceApplicationAmount = 300.0

var (
	ErrInvoiceApplicationNotFound = infraerrors.NotFound(
		"INVOICE_APPLICATION_NOT_FOUND",
		"invoice application not found",
	)
	ErrInvoiceApplicationOrdersRequired = infraerrors.BadRequest(
		"INVOICE_APPLICATION_ORDERS_REQUIRED",
		"at least one recharge order must be selected",
	)
	ErrInvoiceApplicationAmountBelowMinimum = infraerrors.BadRequest(
		"INVOICE_APPLICATION_AMOUNT_BELOW_MINIMUM",
		"selected recharge orders total amount is below the minimum invoice amount",
	)
	ErrInvoiceApplicationOrderNotEligible = infraerrors.BadRequest(
		"INVOICE_APPLICATION_ORDER_NOT_ELIGIBLE",
		"recharge order is not eligible for invoicing",
	)
	ErrInvoiceApplicationOrderAlreadyClaimed = infraerrors.BadRequest(
		"INVOICE_APPLICATION_ORDER_ALREADY_CLAIMED",
		"recharge order has already been claimed by another invoice application",
	)
	ErrInvoiceApplicationInvalidInvoiceType = infraerrors.BadRequest(
		"INVOICE_APPLICATION_INVOICE_TYPE_INVALID",
		"invoice type is invalid",
	)
	ErrInvoiceApplicationInvalidTitleType = infraerrors.BadRequest(
		"INVOICE_APPLICATION_TITLE_TYPE_INVALID",
		"invoice title type is invalid",
	)
	ErrInvoiceApplicationTitleRequired = infraerrors.BadRequest(
		"INVOICE_APPLICATION_TITLE_REQUIRED",
		"invoice title is required",
	)
	ErrInvoiceApplicationTaxNumberRequired = infraerrors.BadRequest(
		"INVOICE_APPLICATION_TAX_NUMBER_REQUIRED",
		"tax number is required for company title",
	)
	ErrInvoiceApplicationBankInfoRequired = infraerrors.BadRequest(
		"INVOICE_APPLICATION_BANK_INFO_REQUIRED",
		"bank name and bank account are required for special VAT invoice",
	)
	ErrInvoiceApplicationEmailInvalid = infraerrors.BadRequest(
		"INVOICE_APPLICATION_EMAIL_INVALID",
		"invoice delivery email is invalid",
	)
	ErrInvoiceApplicationStatusTransitionInvalid = infraerrors.BadRequest(
		"INVOICE_APPLICATION_STATUS_TRANSITION_INVALID",
		"invoice application status cannot be changed from its current state",
	)
	ErrInvoiceApplicationRejectReasonRequired = infraerrors.BadRequest(
		"INVOICE_APPLICATION_REJECT_REASON_REQUIRED",
		"reject reason is required",
	)
	ErrInvoiceApplicationInvoiceInfoRequired = infraerrors.BadRequest(
		"INVOICE_APPLICATION_INVOICE_INFO_REQUIRED",
		"invoice number and invoice amount are required",
	)
	ErrInvoiceApplicationInvoiceAmountInvalid = infraerrors.BadRequest(
		"INVOICE_APPLICATION_INVOICE_AMOUNT_INVALID",
		"invoice amount must be greater than zero",
	)
)

// InvoiceApplicationItem 是一条开票申请占用的一笔充值订单。
//
// ReleasedAt 为空表示该充值记录仍被占用；申请被驳回或取消时写入该时间以释放占用，
// 明细行本身保留以供审计。
type InvoiceApplicationItem struct {
	ID            int64
	ApplicationID int64
	OrderID       int64
	OrderNo       string
	Amount        float64
	Currency      string
	PaidAt        *time.Time
	ReleasedAt    *time.Time
	CreatedAt     time.Time
}

// InvoiceApplication 是一条开票申请。
//
// UserEmail / UserName 与明细中的 OrderNo / Amount / PaidAt 均为快照，
// 不依赖 users 与 payment_orders 的关联查询，保证用户或订单被删除后凭证仍可读。
type InvoiceApplication struct {
	ID     int64
	UserID int64
	// UserEmail 用户邮箱（快照）
	UserEmail string
	// UserName 用户名称（快照）
	UserName string
	// Status 申请状态
	Status string
	// InvoiceType 发票类型
	InvoiceType string
	// TitleType 抬头类型
	TitleType string
	// Title 发票抬头
	Title string
	// TaxNumber 纳税人识别号
	TaxNumber string
	// RegisteredAddress 注册地址
	RegisteredAddress string
	// RegisteredPhone 注册电话
	RegisteredPhone string
	// BankName 开户银行
	BankName string
	// BankAccount 银行账号
	BankAccount string
	// Email 接收发票的邮箱
	Email string
	// Remark 用户备注
	Remark string
	// TotalAmount 申请金额合计
	TotalAmount float64
	// InvoiceAmount 实际开票金额
	InvoiceAmount *float64
	// InvoiceNumber 发票号码
	InvoiceNumber string
	// InvoiceCode 发票代码
	InvoiceCode string
	// IssuedAt 开票日期
	IssuedAt *time.Time
	// ReviewedBy 审核人用户ID
	ReviewedBy *int64
	// ReviewedAt 审核时间
	ReviewedAt *time.Time
	// RejectedReason 驳回原因
	RejectedReason string
	// AdminRemark 管理员备注
	AdminRemark string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	// Items 关联的充值记录明细
	Items []InvoiceApplicationItem
}

// IsTerminal 判断申请是否处于终态（不可再流转）。
func (a *InvoiceApplication) IsTerminal() bool {
	if a == nil {
		return true
	}
	switch a.Status {
	case InvoiceApplicationStatusApproved,
		InvoiceApplicationStatusIssued,
		InvoiceApplicationStatusRejected,
		InvoiceApplicationStatusCancelled:
		return true
	default:
		return false
	}
}
