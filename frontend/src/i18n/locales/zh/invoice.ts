/**
 * 用户端开票申请文案。
 * 命名空间为 `invoice.*`。
 */
export default {
  invoice: {
    title: '开票申请',
    description: '基于充值记录申请开具发票',
    tabEligible: '可开票记录',
    tabMyApplications: '我的申请',

    eligible: {
      empty: '暂无可开票的充值记录',
      orderNo: '订单号',
      paidAt: '支付时间',
      payAmount: '实付金额',
      creditedAmount: '入账金额',
      currency: '币种',
      selectAll: '全选'
    },

    selection: {
      count: '已选 {count} 笔',
      amount: '已选金额',
      minAmount: '最低开票金额',
      needMore: '还需选择 {amount} 才能达到最低开票金额',
      apply: '申请开票',
      noneSelected: '请至少选择一笔充值记录'
    },

    form: {
      title: '申请开票',
      invoiceType: '发票类型',
      titleType: '抬头类型',
      invoiceTitle: '发票抬头',
      taxNumber: '纳税人识别号',
      registeredAddress: '注册地址',
      registeredPhone: '注册电话',
      bankName: '开户银行',
      bankAccount: '银行账号',
      email: '接收邮箱',
      remark: '备注',
      submit: '提交申请',
      cancel: '取消',
      titleRequired: '发票抬头不能为空',
      emailRequired: '接收邮箱不能为空',
      emailInvalid: '接收邮箱格式不正确',
      taxNumberRequired: '企业抬头必须填写纳税人识别号',
      bankInfoRequired: '电子专用发票必须填写开户银行和银行账号'
    },

    invoiceType: {
      electron_normal: '电子普通发票',
      electron_special: '电子专用发票'
    },

    titleType: {
      personal: '个人',
      company: '企业'
    },

    status: {
      pending: '待审核',
      approved: '已审核待开票',
      issued: '已开票',
      rejected: '已驳回',
      cancelled: '已取消'
    },

    applications: {
      empty: '暂无开票申请',
      id: '申请单号',
      invoiceType: '发票类型',
      title: '发票抬头',
      amount: '申请金额',
      createdAt: '申请时间',
      status: '状态',
      invoiceNumber: '发票号码',
      invoiceAmount: '开票金额',
      rejectedReason: '驳回原因',
      viewDetail: '查看详情',
      cancel: '取消申请',
      cancelConfirm: '确定要取消这条开票申请吗？',
      cancelSuccess: '开票申请已取消',
      submitSuccess: '开票申请已提交'
    },

    errors: {
      INVOICE_APPLICATION_NOT_FOUND: '开票申请不存在。',
      INVOICE_APPLICATION_ORDERS_REQUIRED: '请至少选择一笔充值记录。',
      INVOICE_APPLICATION_AMOUNT_BELOW_MINIMUM: '所选充值记录合计低于最低开票金额。',
      INVOICE_APPLICATION_ORDER_NOT_ELIGIBLE: '所选充值记录中存在不可开票的记录。',
      INVOICE_APPLICATION_ORDER_ALREADY_CLAIMED: '所选充值记录已被其他开票申请占用。',
      INVOICE_APPLICATION_INVALID_INVOICE_TYPE: '发票类型无效。',
      INVOICE_APPLICATION_INVALID_TITLE_TYPE: '抬头类型无效。',
      INVOICE_APPLICATION_TITLE_REQUIRED: '发票抬头不能为空。',
      INVOICE_APPLICATION_TAX_NUMBER_REQUIRED: '企业抬头必须填写纳税人识别号。',
      INVOICE_APPLICATION_BANK_INFO_REQUIRED: '电子专用发票必须填写开户银行和银行账号。',
      INVOICE_APPLICATION_EMAIL_INVALID: '接收邮箱格式不正确。',
      INVOICE_APPLICATION_STATUS_TRANSITION_INVALID: '当前状态不允许该操作。',
      INVOICE_APPLICATION_REJECT_REASON_REQUIRED: '驳回原因不能为空。',
      INVOICE_APPLICATION_INVOICE_INFO_REQUIRED: '发票号码和开票金额不能为空。',
      INVOICE_APPLICATION_INVOICE_AMOUNT_INVALID: '开票金额必须大于 0。'
    }
  }
}
