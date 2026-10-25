/**
 * User-facing invoice application strings.
 * Namespaced under `invoice.*`.
 */
export default {
  invoice: {
    title: 'Invoice Applications',
    description: 'Apply for an invoice based on your recharge records',
    tabEligible: 'Eligible Orders',
    tabMyApplications: 'My Applications',

    eligible: {
      empty: 'No invoiceable recharge records',
      orderNo: 'Order No',
      paidAt: 'Paid At',
      payAmount: 'Pay Amount',
      creditedAmount: 'Credited Amount',
      currency: 'Currency',
      selectAll: 'Select all'
    },

    selection: {
      count: '{count} selected',
      amount: 'Selected amount',
      minAmount: 'Minimum amount',
      needMore: 'Select {amount} more to reach the minimum invoice amount',
      apply: 'Apply for Invoice',
      noneSelected: 'Please select at least one recharge order'
    },

    form: {
      title: 'Apply for Invoice',
      invoiceType: 'Invoice Type',
      titleType: 'Title Type',
      invoiceTitle: 'Invoice Title',
      taxNumber: 'Tax ID Number',
      registeredAddress: 'Registered Address',
      registeredPhone: 'Registered Phone',
      bankName: 'Bank Name',
      bankAccount: 'Bank Account',
      email: 'Delivery Email',
      remark: 'Remark',
      submit: 'Submit Application',
      cancel: 'Cancel',
      titleRequired: 'Invoice title is required',
      emailRequired: 'Delivery email is required',
      emailInvalid: 'Delivery email is invalid',
      taxNumberRequired: 'Tax ID number is required for enterprise titles',
      bankInfoRequired: 'Bank name and bank account are required for special VAT invoices'
    },

    invoiceType: {
      electron_normal: 'Electronic Normal Invoice',
      electron_special: 'Electronic Special VAT Invoice'
    },

    titleType: {
      personal: 'Personal',
      company: 'Enterprise'
    },

    status: {
      pending: 'Pending Review',
      approved: 'Approved',
      issued: 'Issued',
      rejected: 'Rejected',
      cancelled: 'Cancelled'
    },

    applications: {
      empty: 'No invoice applications yet',
      id: 'Application ID',
      invoiceType: 'Invoice Type',
      title: 'Invoice Title',
      amount: 'Amount',
      createdAt: 'Applied At',
      status: 'Status',
      invoiceNumber: 'Invoice Number',
      invoiceAmount: 'Invoice Amount',
      rejectedReason: 'Rejection Reason',
      viewDetail: 'View Details',
      cancel: 'Cancel Application',
      cancelConfirm: 'Are you sure you want to cancel this invoice application?',
      cancelSuccess: 'Invoice application cancelled',
      submitSuccess: 'Invoice application submitted'
    },

    errors: {
      INVOICE_APPLICATION_NOT_FOUND: 'Invoice application not found.',
      INVOICE_APPLICATION_ORDERS_REQUIRED: 'Please select at least one recharge order.',
      INVOICE_APPLICATION_AMOUNT_BELOW_MINIMUM: 'The selected recharge orders total is below the minimum invoice amount.',
      INVOICE_APPLICATION_ORDER_NOT_ELIGIBLE: 'One of the selected recharge orders is not eligible for invoicing.',
      INVOICE_APPLICATION_ORDER_ALREADY_CLAIMED: 'One of the selected recharge orders has already been claimed by another invoice application.',
      INVOICE_APPLICATION_INVALID_INVOICE_TYPE: 'Invalid invoice type.',
      INVOICE_APPLICATION_INVALID_TITLE_TYPE: 'Invalid invoice title type.',
      INVOICE_APPLICATION_TITLE_REQUIRED: 'Invoice title is required.',
      INVOICE_APPLICATION_TAX_NUMBER_REQUIRED: 'Tax ID number is required for enterprise titles.',
      INVOICE_APPLICATION_BANK_INFO_REQUIRED: 'Bank name and bank account are required for special VAT invoices.',
      INVOICE_APPLICATION_EMAIL_INVALID: 'Delivery email is invalid.',
      INVOICE_APPLICATION_STATUS_TRANSITION_INVALID: 'The invoice application status cannot be changed from its current state.',
      INVOICE_APPLICATION_REJECT_REASON_REQUIRED: 'Rejection reason is required.',
      INVOICE_APPLICATION_INVOICE_INFO_REQUIRED: 'Invoice number and invoice amount are required.',
      INVOICE_APPLICATION_INVOICE_AMOUNT_INVALID: 'Invoice amount must be greater than zero.'
    }
  }
}
