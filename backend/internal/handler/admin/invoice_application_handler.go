package admin

import (
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

// InvoiceApplicationHandler 处理管理端的开票申请审核与开票登记。
type InvoiceApplicationHandler struct {
	invoiceService *service.InvoiceApplicationService
}

// NewInvoiceApplicationHandler 创建管理端开票申请 handler。
func NewInvoiceApplicationHandler(invoiceService *service.InvoiceApplicationService) *InvoiceApplicationHandler {
	return &InvoiceApplicationHandler{invoiceService: invoiceService}
}

// ListApplications 分页查询全部开票申请。
// GET /api/v1/admin/invoice/applications
func (h *InvoiceApplicationHandler) ListApplications(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)

	var userID int64
	if raw := c.Query("user_id"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			userID = parsed
		}
	}

	apps, result, err := h.invoiceService.AdminList(
		c.Request.Context(),
		pagination.PaginationParams{Page: page, PageSize: pageSize},
		service.InvoiceApplicationListFilters{
			Status:  strings.TrimSpace(c.Query("status")),
			UserID:  userID,
			Keyword: strings.TrimSpace(c.Query("keyword")),
		},
	)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Paginated(c, dto.InvoiceApplicationsFromService(apps), result.Total, result.Page, result.PageSize)
}

// GetApplication 查询开票申请详情。
// GET /api/v1/admin/invoice/applications/:id
func (h *InvoiceApplicationHandler) GetApplication(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid invoice application ID")
		return
	}

	app, err := h.invoiceService.AdminGet(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, dto.InvoiceApplicationFromService(app))
}

// ApproveApplication 审核通过一条待审核申请。
// POST /api/v1/admin/invoice/applications/:id/approve
func (h *InvoiceApplicationHandler) ApproveApplication(c *gin.Context) {
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

	if err := h.invoiceService.AdminApprove(c.Request.Context(), id, subject.UserID); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "invoice application approved"})
}

// RejectApplicationRequest 是驳回开票申请的请求体。
type RejectApplicationRequest struct {
	// Reason 驳回原因（必填）
	Reason string `json:"reason"`
	// Remark 管理员备注
	Remark string `json:"remark"`
}

// RejectApplication 驳回一条待审核申请并释放其占用的充值记录。
// POST /api/v1/admin/invoice/applications/:id/reject
func (h *InvoiceApplicationHandler) RejectApplication(c *gin.Context) {
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

	var req RejectApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.invoiceService.AdminReject(c.Request.Context(), id, subject.UserID, req.Reason, req.Remark); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "invoice application rejected"})
}

// IssueApplicationRequest 是登记实际开票信息的请求体。
type IssueApplicationRequest struct {
	// InvoiceNumber 发票号码（必填）
	InvoiceNumber string `json:"invoice_number"`
	// InvoiceCode 发票代码
	InvoiceCode string `json:"invoice_code"`
	// InvoiceAmount 开票金额
	InvoiceAmount float64 `json:"invoice_amount"`
	// IssuedAt 开票日期
	IssuedAt *time.Time `json:"issued_at"`
	// AdminRemark 管理员备注
	AdminRemark string `json:"admin_remark"`
}

// IssueApplication 在审核通过的申请上登记实际开具的发票信息。
// POST /api/v1/admin/invoice/applications/:id/issue
func (h *InvoiceApplicationHandler) IssueApplication(c *gin.Context) {
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

	var req IssueApplicationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request body")
		return
	}

	if err := h.invoiceService.AdminIssue(c.Request.Context(), id, subject.UserID, service.IssueInvoiceApplicationInput{
		InvoiceNumber: req.InvoiceNumber,
		InvoiceCode:   req.InvoiceCode,
		InvoiceAmount: req.InvoiceAmount,
		IssuedAt:      req.IssuedAt,
		AdminRemark:   req.AdminRemark,
	}); err != nil {
		response.ErrorFrom(c, err)
		return
	}

	response.Success(c, gin.H{"message": "invoice application issued"})
}
