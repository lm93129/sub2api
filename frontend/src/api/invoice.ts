/**
 * Invoice Application API endpoints
 * Handles invoice (开票申请) operations for regular users and administrators
 */

import { apiClient } from './client'
import type {
  BasePaginationResponse,
  CreateInvoiceApplicationRequest,
  InvoiceApplication,
  InvoiceApplicationItem,
  InvoiceableOrder,
  IssueInvoiceApplicationRequest,
  RejectInvoiceApplicationRequest
} from '@/types'

/** Minimum amount (CNY) required for a single invoice application */
export const MIN_INVOICE_AMOUNT = 300

export const invoiceAPI = {
  /** Get recharge orders that are eligible for invoicing */
  getEligibleOrders(params?: { page?: number; page_size?: number }) {
    return apiClient.get<BasePaginationResponse<InvoiceableOrder>>('/invoice/eligible-orders', { params })
  },

  /** Get current user's invoice applications */
  getMyApplications(params?: { page?: number; page_size?: number; status?: string }) {
    return apiClient.get<BasePaginationResponse<InvoiceApplication>>('/invoice/applications', { params })
  },

  /** Get a single invoice application */
  getApplication(id: number) {
    return apiClient.get<InvoiceApplication>(`/invoice/applications/${id}`)
  },

  /** Create a new invoice application */
  createApplication(data: CreateInvoiceApplicationRequest) {
    return apiClient.post<InvoiceApplication>('/invoice/applications', data)
  },

  /** Cancel a pending invoice application */
  cancelApplication(id: number) {
    return apiClient.post<InvoiceApplication>(`/invoice/applications/${id}/cancel`)
  }
}

export const adminInvoiceAPI = {
  /** Get all invoice applications (paginated, with filters) */
  getApplications(params?: {
    page?: number
    page_size?: number
    status?: string
    user_id?: number
    keyword?: string
  }) {
    return apiClient.get<BasePaginationResponse<InvoiceApplication>>('/admin/invoice/applications', { params })
  },

  /** Get a single invoice application */
  getApplication(id: number) {
    return apiClient.get<InvoiceApplication>(`/admin/invoice/applications/${id}`)
  },

  /** Approve a pending application */
  approve(id: number) {
    return apiClient.post<{ message: string }>(`/admin/invoice/applications/${id}/approve`)
  },

  /** Reject a pending application (reason is required) */
  reject(id: number, data: RejectInvoiceApplicationRequest) {
    return apiClient.post<{ message: string }>(`/admin/invoice/applications/${id}/reject`, data)
  },

  /** Register the issued invoice information on an approved application */
  issue(id: number, data: IssueInvoiceApplicationRequest) {
    return apiClient.post<{ message: string }>(`/admin/invoice/applications/${id}/issue`, data)
  }
}

export type { InvoiceApplication, InvoiceApplicationItem, InvoiceableOrder }
