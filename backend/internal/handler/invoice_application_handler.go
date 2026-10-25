package handler

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// InvoiceApplicationHandler 处理用户侧的开票申请。
type InvoiceApplicationHandler struct {
	invoiceService *service.InvoiceApplicationService
}

// NewInvoiceApplicationHandler 创建用户端开票申请 handler。
func NewInvoiceApplicationHandler(invoiceService *service.InvoiceApplicationService) *InvoiceApplicationHandler {
	return &InvoiceApplicationHandler{invoiceService: invoiceService}
}

// ListEligibleOrders 列出当前用户可申请开票的充值记录。
// GET /api/v1/invoice/eligible-orders
func (h *InvoiceApplicationHandler) ListEligibleOrders(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	page, pageSize := response.ParsePagination(c)
	orders, result, err := h.invoiceService.ListEligibleOrders(
		c.Request.Context(),
		subject.UserID,
		pagination.PaginationParams{Page: page, PageSize: pageSize},
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, dto.InvoiceApplicationOrdersFromService(orders), result.Total, result.Page, result.PageSize)
}

// ListApplications 分页查询当前用户的开票申请。
// GET /api/v1/invoice/applications
func (h *InvoiceApplicationHandler) ListApplications(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	page, pageSize := response.ParsePagination(c)
	apps, result, err := h.invoiceService.ListUserApplications(
		c.Request.Context(),
		subject.UserID,
		pagination.PaginationParams{Page: page, PageSize: pageSize},
		service.InvoiceApplicationListFilters{Status: c.Query("status")},
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, dto.InvoiceApplicationsFromService(apps), result.Total, result.Page, result.PageSize)
}

// GetApplication 查询当前用户的开票申请详情。
// GET /api/v1/invoice/applications/:id
func (h *InvoiceApplicationHandler) GetApplication(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid invoice application ID")
		return
	}

	app, err := h.invoiceService.GetUserApplication(c.Request.Context(), subject.UserID, id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.InvoiceApplicationFromService(app))
}

// CreateApplicationRequest 是提交开票申请的请求体。
type CreateApplicationRequest struct {
	// OrderIDs 选中的充值订单 ID
	OrderIDs []int64 `json:"order_ids" binding:"required"`
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
	// Remark 备注
	Remark string `json:"remark"`
}

// CreateApplication 提交开票申请。
// POST /api/v1/invoice/applications
func (h *InvoiceApplicationHandler) CreateApplication(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	var req CreateApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	app, err := h.invoiceService.CreateApplication(c.Request.Context(), subject.UserID, service.CreateInvoiceApplicationInput{
		OrderIDs:           req.OrderIDs,
		InvoiceType:        req.InvoiceType,
		TitleType:          req.TitleType,
		Title:              req.Title,
		TaxNumber:          req.TaxNumber,
		RegisteredAddress:  req.RegisteredAddress,
		RegisteredPhone:    req.RegisteredPhone,
		BankName:           req.BankName,
		BankAccount:        req.BankAccount,
		Email:              req.Email,
		Remark:             req.Remark,
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.InvoiceApplicationFromService(app))
}

// CancelApplication 取消当前用户的待审核开票申请。
// POST /api/v1/invoice/applications/:id/cancel
func (h *InvoiceApplicationHandler) CancelApplication(c *gin.Context) {
	subject, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}

	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid invoice application ID")
		return
	}

	if err := h.invoiceService.CancelApplication(c.Request.Context(), subject.UserID, id); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "invoice application cancelled"})
}
