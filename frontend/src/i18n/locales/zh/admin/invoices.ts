/**
 * 管理端开票申请文案。
 * 命名空间为 `admin.invoices.*`。
 */
export default {
  invoices: {
    title: '开票申请',
    description: '审核开票申请并登记实际开具的发票信息',
    searchPlaceholder: '按抬头、税号、邮箱或用户名搜索',
    allStatuses: '全部状态',
    empty: '暂无开票申请',

    columns: {
      user: '用户',
      title: '发票抬头',
      amount: '金额',
      status: '状态',
      createdAt: '申请时间',
      actions: '操作'
    },

    status: {
      pending: '待审核',
      approved: '已审核待开票',
      issued: '已开票',
      rejected: '已驳回',
      cancelled: '已取消'
    },

    invoiceType: {
      electron_normal: '电子普通发票',
      electron_special: '电子专用发票'
    },

    titleType: {
      personal: '个人',
      company: '企业'
    },

    actions: {
      view: '查看详情',
      approve: '审核通过',
      reject: '驳回',
      issue: '登记开票'
    },

    detail: {
      title: '开票申请详情',
      applicationId: '申请单号',
      user: '用户',
      userEmail: '邮箱',
      userName: '用户名',
      status: '状态',
      invoiceType: '发票类型',
      titleType: '抬头类型',
      invoiceTitle: '发票抬头',
      taxNumber: '纳税人识别号',
      registeredAddress: '注册地址',
      registeredPhone: '注册电话',
      bankName: '开户银行',
      bankAccount: '银行账号',
      email: '接收邮箱',
      remark: '用户备注',
      totalAmount: '申请金额',
      invoiceAmount: '开票金额',
      invoiceNumber: '发票号码',
      invoiceCode: '发票代码',
      issuedAt: '开票日期',
      reviewedAt: '审核时间',
      rejectedReason: '驳回原因',
      adminRemark: '管理员备注',
      createdAt: '申请时间',
      updatedAt: '更新时间'
    },

    items: {
      title: '关联充值记录',
      orderNo: '订单号',
      amount: '金额',
      paidAt: '支付时间',
      releasedAt: '释放时间',
      empty: '暂无关联的充值记录'
    },

    approve: {
      title: '审核通过开票申请',
      confirm: '确定审核通过这条开票申请吗？',
      success: '开票申请已审核通过'
    },

    reject: {
      title: '驳回开票申请',
      reason: '驳回原因',
      reasonPlaceholder: '请填写驳回原因',
      remark: '管理员备注',
      submit: '驳回',
      success: '开票申请已驳回',
      reasonRequired: '驳回原因不能为空'
    },

    issue: {
      title: '登记开票信息',
      invoiceNumber: '发票号码',
      invoiceCode: '发票代码',
      invoiceAmount: '开票金额',
      issuedAt: '开票日期',
      adminRemark: '管理员备注',
      submit: '登记',
      success: '开票信息已登记',
      invoiceNumberRequired: '发票号码不能为空',
      invoiceAmountRequired: '开票金额必须大于 0'
    }
  }
}
