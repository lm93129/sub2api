/**
 * Admin invoice application strings.
 * Namespaced under `admin.invoices.*`.
 */
export default {
  invoices: {
    title: 'Invoice Applications',
    description: 'Review invoice applications and register issued invoices',
    searchPlaceholder: 'Search by title, tax number, email or username',
    allStatuses: 'All Statuses',
    empty: 'No invoice applications found',

    columns: {
      user: 'User',
      title: 'Title',
      amount: 'Amount',
      status: 'Status',
      createdAt: 'Applied At',
      actions: 'Actions'
    },

    status: {
      pending: 'Pending Review',
      approved: 'Approved',
      issued: 'Issued',
      rejected: 'Rejected',
      cancelled: 'Cancelled'
    },

    invoiceType: {
      electron_normal: 'Electronic Normal Invoice',
      electron_special: 'Electronic Special VAT Invoice'
    },

    titleType: {
      personal: 'Personal',
      company: 'Enterprise'
    },

    actions: {
      view: 'View',
      approve: 'Approve',
      reject: 'Reject',
      issue: 'Register Invoice'
    },

    detail: {
      title: 'Invoice Application Detail',
      applicationId: 'Application ID',
      user: 'User',
      userEmail: 'Email',
      userName: 'Username',
      status: 'Status',
      invoiceType: 'Invoice Type',
      titleType: 'Title Type',
      invoiceTitle: 'Invoice Title',
      taxNumber: 'Tax ID Number',
      registeredAddress: 'Registered Address',
      registeredPhone: 'Registered Phone',
      bankName: 'Bank Name',
      bankAccount: 'Bank Account',
      email: 'Delivery Email',
      remark: 'User Remark',
      totalAmount: 'Applied Amount',
      invoiceAmount: 'Invoice Amount',
      invoiceNumber: 'Invoice Number',
      invoiceCode: 'Invoice Code',
      issuedAt: 'Issued At',
      reviewedAt: 'Reviewed At',
      rejectedReason: 'Rejection Reason',
      adminRemark: 'Admin Remark',
      createdAt: 'Applied At',
      updatedAt: 'Updated At'
    },

    items: {
      title: 'Recharge Records',
      orderNo: 'Order No',
      amount: 'Amount',
      paidAt: 'Paid At',
      releasedAt: 'Released At',
      empty: 'No recharge records linked'
    },

    approve: {
      title: 'Approve Invoice Application',
      confirm: 'Approve this invoice application?',
      success: 'Invoice application approved'
    },

    reject: {
      title: 'Reject Invoice Application',
      reason: 'Rejection Reason',
      reasonPlaceholder: 'Please enter the rejection reason',
      remark: 'Admin Remark',
      submit: 'Reject',
      success: 'Invoice application rejected',
      reasonRequired: 'Rejection reason is required'
    },

    issue: {
      title: 'Register Issued Invoice',
      invoiceNumber: 'Invoice Number',
      invoiceCode: 'Invoice Code',
      invoiceAmount: 'Invoice Amount',
      issuedAt: 'Issued At',
      adminRemark: 'Admin Remark',
      submit: 'Register',
      success: 'Invoice registered',
      invoiceNumberRequired: 'Invoice number is required',
      invoiceAmountRequired: 'Invoice amount must be greater than zero'
    }
  }
}
