package adminaccess

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestInvoiceApplicationRoutesRegistered 验证开票申请的管理端接口都已登记到权限目录。
//
// 目录按「HTTP 方法 + Gin 路由模式」精确匹配（中间件传入的是 c.FullPath()，
// 即含 :id 占位符的注册路径），不会继承同目录的其他权限；
// 未登记的接口对受限管理密钥直接拒绝，因此新增管理端接口必须在此登记。
func TestInvoiceApplicationRoutesRegistered(t *testing.T) {
	cases := []struct {
		permission string
		method     string
		path       string
	}{
		{"invoice.read", "GET", "/api/v1/admin/invoice/applications"},
		{"invoice.read", "GET", "/api/v1/admin/invoice/applications/:id"},
		{"invoice.review", "POST", "/api/v1/admin/invoice/applications/:id/approve"},
		{"invoice.review", "POST", "/api/v1/admin/invoice/applications/:id/reject"},
		{"invoice.issue", "POST", "/api/v1/admin/invoice/applications/:id/issue"},
	}

	for _, tc := range cases {
		t.Run(tc.permission+" "+tc.method+" "+tc.path, func(t *testing.T) {
			policy := Policy{GlobalPermissions: []string{tc.permission}}
			require.True(t, RouteAllowed(policy, tc.method, tc.path),
				"接口 %s %s 未登记到权限目录", tc.method, tc.path)

			// 未授予该权限时必须拒绝。
			require.False(t, RouteAllowed(Policy{}, tc.method, tc.path),
				"未授权的管理密钥不应被允许访问 %s %s", tc.method, tc.path)
		})
	}
}

// TestInvoiceApplicationPermissionsAreDistinct 验证三个开票权限彼此独立。
//
// 审核权限不能顺带获得登记开票信息的权限，否则受限管理密钥的授权边界会被放大。
func TestInvoiceApplicationPermissionsAreDistinct(t *testing.T) {
	readOnly := Policy{GlobalPermissions: []string{"invoice.read"}}
	require.False(t, RouteAllowed(readOnly, "POST", "/api/v1/admin/invoice/applications/:id/approve"))
	require.False(t, RouteAllowed(readOnly, "POST", "/api/v1/admin/invoice/applications/:id/issue"))

	reviewOnly := Policy{GlobalPermissions: []string{"invoice.review"}}
	require.False(t, RouteAllowed(reviewOnly, "POST", "/api/v1/admin/invoice/applications/:id/issue"))
	require.True(t, RouteAllowed(reviewOnly, "POST", "/api/v1/admin/invoice/applications/:id/approve"))
	require.True(t, RouteAllowed(reviewOnly, "POST", "/api/v1/admin/invoice/applications/:id/reject"))

	issueOnly := Policy{GlobalPermissions: []string{"invoice.issue"}}
	require.True(t, RouteAllowed(issueOnly, "POST", "/api/v1/admin/invoice/applications/:id/issue"))
	require.False(t, RouteAllowed(issueOnly, "POST", "/api/v1/admin/invoice/applications/:id/approve"))
}
