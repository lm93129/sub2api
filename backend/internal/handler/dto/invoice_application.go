package dto

import (
	"time"

	"github.com/Wei-Shaw/sub2api/internal/service"
)

// InvoiceApplicationOrder 是可开票的充值记录。
type InvoiceApplicationOrder struct {
	OrderID int64 `json:"order_id"`
	// OrderNo 充值订单号
	OrderNo string `json:"order_no"`
	// PayAmount 实付金额，开票金额的计算基数
	PayAmount float64 `json:"pay_amount"`
	// TotalAmount 入账金额（含赠送额度）
	TotalAmount float64 `json:"total_amount"`
	// Currency 支付币种
	Currency string `json:"currency"`
	// PaidAt 支付时间
	PaidAt *time.Time `json:"paid_at,omitempty"`
	// CreatedAt 充值下单时间
	CreatedAt time.Time `json:"created_at"`
}

// InvoiceApplicationItem 是申请占用的���笔充值记录。
type InvoiceApplicationItem struct {
	ID            int64 `json:"id"`
	OrderID       int64 `json:"order_id"`
	OrderNo       string `json:"order_no"`
	Amount        float64 `json:"amount"`
	Currency      string `json:"currency"`
	PaidAt        *time.Time `json:"paid_at,omitempty"`
	ReleasedAt    *time.Time `json:"released_at,omitempty"`
}

// InvoiceApplication 是开票申请。
type InvoiceApplication struct {
	ID       int64 `json:"id"`
	UserID   int64 `json:"user_id"`
	// UserEmail 申请用户邮箱
	UserEmail string `json:"user_email"`
	// UserName 申请用户名称
	UserName string `json:"user_name"`
	// Status 申请状态
	Status string `json:"status"`
	// InvoiceType 发票类型
	InvoiceType string `json:"invoice_type"`
	// TitleType 抬头类型
	TitleType string `json:"title_type"`
	// Title 发票抬头
	Title string `json:"title"`
	// TaxNumber 纳税人识别号
	TaxNumber string `json:"tax_number"`
	// RegisteredAddress 注册地址
	RegisteredAddress string `json:"registered_address"`
	// RegisteredPhone 注册电话
	RegisteredPhone string `json:"registered_phone"`
	// BankName 开户银行
	BankName string `json:"bank_name"`
	// BankAccount 银行账号
	BankAccount string `json:"bank_account"`
	// Email 接收发票的邮箱
	Email string `json:"email"`
	// Remark 用户备注
	Remark string `json:"remark"`
	// TotalAmount 申请金额合计
	TotalAmount float64 `json:"total_amount"`
	// InvoiceAmount 实际开票金额
	InvoiceAmount *float64 `json:"invoice_amount,omitempty"`
	// InvoiceNumber 发票号码
	InvoiceNumber string `json:"invoice_number"`
	// InvoiceCode 发票代码
	InvoiceCode string `json:"invoice_code"`
	// IssuedAt 开票日期
	IssuedAt *time.Time `json:"issued_at,omitempty"`
	// ReviewedBy 审核人用户ID
	ReviewedBy *int64 `json:"reviewed_by,omitempty"`
	// ReviewedAt 审核时间
	ReviewedAt *time.Time `json:"reviewed_at,omitempty"`
	// RejectedReason 驳回原因
	RejectedReason string `json:"rejected_reason"`
	// AdminRemark 管理员备注
	AdminRemark string `json:"admin_remark"`
	// CreatedAt 申请时间
	CreatedAt time.Time `json:"created_at"`
	// UpdatedAt 更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// Items 关联的充值记录明细
	Items []InvoiceApplicationItem `json:"items"`
}

// InvoiceApplicationFromService 把 service 层结构体转换为 DTO。
func InvoiceApplicationFromService(app *service.InvoiceApplication) *InvoiceApplication {
	if app == nil {
		return nil
	}
	out := &InvoiceApplication{
		ID:                app.ID,
		UserID:            app.UserID,
		UserEmail:         app.UserEmail,
		UserName:          app.UserName,
		Status:            app.Status,
		InvoiceType:       app.InvoiceType,
		TitleType:         app.TitleType,
		Title:             app.Title,
		TaxNumber:         app.TaxNumber,
		RegisteredAddress: app.RegisteredAddress,
		RegisteredPhone:   app.RegisteredPhone,
		BankName:          app.BankName,
		BankAccount:       app.BankAccount,
		Email:             app.Email,
		Remark:            app.Remark,
		TotalAmount:       app.TotalAmount,
		InvoiceNumber:     app.InvoiceNumber,
		InvoiceCode:       app.InvoiceCode,
		RejectedReason:    app.RejectedReason,
		AdminRemark:       app.AdminRemark,
		CreatedAt:         app.CreatedAt,
		UpdatedAt:         app.UpdatedAt,
		Items:             make([]InvoiceApplicationItem, 0, len(app.Items)),
	}
	if app.InvoiceAmount != nil {
		amount := *app.InvoiceAmount
		out.InvoiceAmount = &amount
	}
	if app.IssuedAt != nil {
		issuedAt := *app.IssuedAt
		out.IssuedAt = &issuedAt
	}
	if app.ReviewedBy != nil {
		reviewedBy := *app.ReviewedBy
		out.ReviewedBy = &reviewedBy
	}
	if app.ReviewedAt != nil {
		reviewedAt := *app.ReviewedAt
		out.ReviewedAt = &reviewedAt
	}
	for i := range app.Items {
		item := &app.Items[i]
		out.Items = append(out.Items, InvoiceApplicationItem{
			ID:         item.ID,
			OrderID:    item.OrderID,
			OrderNo:    item.OrderNo,
			Amount:     item.Amount,
			Currency:   item.Currency,
			PaidAt:     item.PaidAt,
			ReleasedAt: item.ReleasedAt,
		})
	}
	return out
}

// InvoiceApplicationsFromService 批量转换开票申请。
func InvoiceApplicationsFromService(apps []service.InvoiceApplication) []InvoiceApplication {
	out := make([]InvoiceApplication, 0, len(apps))
	for i := range apps {
		converted := InvoiceApplicationFromService(&apps[i])
		if converted != nil {
			out = append(out, *converted)
		}
	}
	return out
}

// InvoiceApplicationOrderFromService 把可开票充值记录转换为 DTO。
func InvoiceApplicationOrderFromService(o *service.InvoiceApplicationOrder) InvoiceApplicationOrder {
	if o == nil {
		return InvoiceApplicationOrder{}
	}
	return InvoiceApplicationOrder{
		OrderID:     o.OrderID,
		OrderNo:     o.OrderNo,
		PayAmount:   o.PayAmount,
		TotalAmount: o.TotalAmount,
		Currency:    o.Currency,
		PaidAt:      o.PaidAt,
		CreatedAt:   o.CreatedAt,
	}
}

// InvoiceApplicationOrdersFromService 批量转换可开票充值记录。
func InvoiceApplicationOrdersFromService(orders []service.InvoiceApplicationOrder) []InvoiceApplicationOrder {
	out := make([]InvoiceApplicationOrder, 0, len(orders))
	for i := range orders {
		out = append(out, InvoiceApplicationOrderFromService(&orders[i]))
	}
	return out
}
