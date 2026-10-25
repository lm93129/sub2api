<template>
  <AppLayout>
    <TablePageLayout>
      <!-- Filters -->
      <template #filters>
        <div class="flex flex-wrap items-center gap-3">
          <div class="flex-1 sm:max-w-72">
            <input
              v-model="keyword"
              type="text"
              class="input"
              :placeholder="t('admin.invoices.searchPlaceholder')"
              @input="handleSearch"
            />
          </div>
          <Select
            v-model="filters.status"
            :options="statusFilterOptions"
            class="w-40"
            @change="handleStatusChange"
          />
          <div class="flex flex-1 flex-wrap items-center justify-end gap-2">
            <button
              @click="loadApplications"
              :disabled="loading"
              class="btn btn-secondary"
              :title="t('common.refresh')"
            >
              <Icon name="refresh" size="md" :class="loading ? 'animate-spin' : ''" />
            </button>
          </div>
        </div>
      </template>

      <!-- Table -->
      <template #table>
        <DataTable :columns="columns" :data="applications" :loading="loading">
          <template #cell-user="{ row }">
            <div class="flex flex-col">
              <span class="text-sm font-medium text-gray-900 dark:text-white">
                {{ row.user_name || row.user_email }}
              </span>
              <span class="text-xs text-gray-500 dark:text-gray-400">{{ row.user_email }}</span>
            </div>
          </template>

          <template #cell-title="{ row }">
            <div class="flex flex-col">
              <span class="text-sm text-gray-900 dark:text-white">{{ row.title }}</span>
              <span class="text-xs text-gray-500 dark:text-gray-400">
                {{ row.tax_number || '-' }}
              </span>
            </div>
          </template>

          <template #cell-total_amount="{ value }">
            <span class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ formatAmount(value) }}
            </span>
          </template>

          <template #cell-status="{ value }">
            <span :class="['badge', statusBadgeClass(value)]">{{ statusLabel(value) }}</span>
          </template>

          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(value) }}</span>
          </template>

          <template #cell-actions="{ row }">
            <div class="flex items-center gap-1">
              <button
                class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-600"
                :title="t('admin.invoices.actions.view')"
                @click="openDetail(row)"
              >
                <Icon name="eye" size="sm" />
                {{ t('admin.invoices.actions.view') }}
              </button>
              <button
                v-if="row.status === 'PENDING'"
                class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-green-600 hover:bg-green-50 dark:text-green-400 dark:hover:bg-green-900/20"
                :title="t('admin.invoices.actions.approve')"
                @click="approveTarget = row"
              >
                <Icon name="check" size="sm" />
                {{ t('admin.invoices.actions.approve') }}
              </button>
              <button
                v-if="row.status === 'PENDING'"
                class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-red-600 hover:bg-red-50 dark:text-red-400 dark:hover:bg-red-900/20"
                :title="t('admin.invoices.actions.reject')"
                @click="openRejectDialog(row)"
              >
                <Icon name="x" size="sm" />
                {{ t('admin.invoices.actions.reject') }}
              </button>
              <button
                v-if="row.status === 'APPROVED'"
                class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-blue-600 hover:bg-blue-50 dark:text-blue-400 dark:hover:bg-blue-900/20"
                :title="t('admin.invoices.actions.issue')"
                @click="openIssueDialog(row)"
              >
                <Icon name="document" size="sm" />
                {{ t('admin.invoices.actions.issue') }}
              </button>
            </div>
          </template>

          <template #empty>
            <EmptyState :title="t('admin.invoices.empty')" />
          </template>
        </DataTable>
      </template>

      <!-- Pagination -->
      <template #pagination>
        <Pagination
          v-if="pagination.total > 0"
          :page="pagination.page"
          :total="pagination.total"
          :page-size="pagination.page_size"
          @update:page="handlePageChange"
          @update:pageSize="handlePageSizeChange"
        />
      </template>
    </TablePageLayout>

    <!-- Detail dialog -->
    <BaseDialog
      :show="showDetailDialog"
      :title="t('admin.invoices.detail.title')"
      width="wide"
      @close="closeDetailDialog"
    >
      <div v-if="detailApplication" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.applicationId') }}</p>
            <p class="font-mono text-sm font-medium text-gray-900 dark:text-white">#{{ detailApplication.id }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.status') }}</p>
            <span :class="['badge', statusBadgeClass(detailApplication.status)]">
              {{ statusLabel(detailApplication.status) }}
            </span>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.userName') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ detailApplication.user_name || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.userEmail') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ detailApplication.user_email || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.invoiceType') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ invoiceTypeLabel(detailApplication.invoice_type) }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.titleType') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ titleTypeLabel(detailApplication.title_type) }}</p>
          </div>
          <div class="col-span-2">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.invoiceTitle') }}</p>
            <p class="text-sm font-medium text-gray-900 dark:text-white">{{ detailApplication.title || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.taxNumber') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ detailApplication.tax_number || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.email') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ detailApplication.email || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.registeredAddress') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ detailApplication.registered_address || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.registeredPhone') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ detailApplication.registered_phone || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.bankName') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ detailApplication.bank_name || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.bankAccount') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ detailApplication.bank_account || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.totalAmount') }}</p>
            <p class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ formatAmount(detailApplication.total_amount) }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.invoiceAmount') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">
              {{ detailApplication.invoice_amount != null ? formatAmount(detailApplication.invoice_amount) : '-' }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.invoiceNumber') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ detailApplication.invoice_number || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.invoiceCode') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ detailApplication.invoice_code || '-' }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.issuedAt') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ detailApplication.issued_at ? formatDateTime(detailApplication.issued_at) : '-' }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.reviewedAt') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">
              {{ detailApplication.reviewed_at ? formatDateTime(detailApplication.reviewed_at) : '-' }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.createdAt') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(detailApplication.created_at) }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.updatedAt') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(detailApplication.updated_at) }}</p>
          </div>
          <div v-if="detailApplication.remark" class="col-span-2">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.remark') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ detailApplication.remark }}</p>
          </div>
          <div v-if="detailApplication.admin_remark" class="col-span-2">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.adminRemark') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ detailApplication.admin_remark }}</p>
          </div>
          <div v-if="detailApplication.rejected_reason" class="col-span-2">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('admin.invoices.detail.rejectedReason') }}</p>
            <p class="text-sm text-red-600 dark:text-red-400">{{ detailApplication.rejected_reason }}</p>
          </div>
        </div>

        <!-- Linked recharge records -->
        <div class="border-t border-gray-200 pt-4 dark:border-dark-600">
          <p class="mb-2 text-xs font-medium text-gray-500 dark:text-gray-400">
            {{ t('admin.invoices.items.title') }}
          </p>
          <div v-if="detailApplication.items?.length" class="overflow-hidden rounded-lg border border-gray-200 dark:border-dark-600">
            <table class="w-full text-left text-sm">
              <thead class="bg-gray-50 text-xs text-gray-500 dark:bg-dark-800 dark:text-gray-400">
                <tr>
                  <th class="px-3 py-2 font-medium">{{ t('admin.invoices.items.orderNo') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('admin.invoices.items.amount') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('admin.invoices.items.paidAt') }}</th>
                  <th class="px-3 py-2 font-medium">{{ t('admin.invoices.items.releasedAt') }}</th>
                </tr>
              </thead>
              <tbody class="divide-y divide-gray-100 dark:divide-dark-700">
                <tr v-for="item in detailApplication.items" :key="item.id">
                  <td class="px-3 py-2 font-mono text-xs text-gray-700 dark:text-gray-300">{{ item.order_no }}</td>
                  <td class="px-3 py-2 text-gray-900 dark:text-white">{{ formatAmount(item.amount, item.currency) }}</td>
                  <td class="px-3 py-2 text-gray-700 dark:text-gray-300">
                    {{ item.paid_at ? formatDateTime(item.paid_at) : '-' }}
                  </td>
                  <td class="px-3 py-2 text-gray-700 dark:text-gray-300">
                    {{ item.released_at ? formatDateTime(item.released_at) : '-' }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
          <p v-else class="text-sm text-gray-500 dark:text-gray-400">{{ t('admin.invoices.items.empty') }}</p>
        </div>
      </div>
    </BaseDialog>

    <!-- Approve confirmation -->
    <ConfirmDialog
      :show="!!approveTarget"
      :title="t('admin.invoices.approve.title')"
      :message="t('admin.invoices.approve.confirm')"
      :confirm-text="t('admin.invoices.actions.approve')"
      :cancel-text="t('common.cancel')"
      @confirm="confirmApprove"
      @cancel="approveTarget = null"
    />

    <!-- Reject dialog -->
    <BaseDialog
      :show="showRejectDialog"
      :title="t('admin.invoices.reject.title')"
      width="normal"
      @close="closeRejectDialog"
    >
      <form id="invoice-reject-form" class="space-y-4" @submit.prevent="submitReject">
        <div>
          <label class="input-label" for="invoice-reject-reason">{{ t('admin.invoices.reject.reason') }} *</label>
          <textarea
            id="invoice-reject-reason"
            v-model="rejectForm.reason"
            rows="3"
            class="input mt-1 w-full"
            :class="rejectErrors.reason ? 'input-error' : ''"
            :placeholder="t('admin.invoices.reject.reasonPlaceholder')"
            required
          />
          <p v-if="rejectErrors.reason" class="input-error-text mt-1">{{ rejectErrors.reason }}</p>
        </div>
        <div>
          <label class="input-label" for="invoice-reject-remark">{{ t('admin.invoices.reject.remark') }}</label>
          <textarea id="invoice-reject-remark" v-model="rejectForm.remark" rows="2" class="input mt-1 w-full" />
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeRejectDialog">
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            form="invoice-reject-form"
            class="btn btn-danger"
            :disabled="submitting"
          >
            {{ submitting ? t('common.processing') : t('admin.invoices.reject.submit') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Issue dialog -->
    <BaseDialog
      :show="showIssueDialog"
      :title="t('admin.invoices.issue.title')"
      width="normal"
      @close="closeIssueDialog"
    >
      <form id="invoice-issue-form" class="space-y-4" @submit.prevent="submitIssue">
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label" for="invoice-issue-number">{{ t('admin.invoices.issue.invoiceNumber') }} *</label>
            <input
              id="invoice-issue-number"
              v-model="issueForm.invoice_number"
              type="text"
              class="input mt-1 w-full"
              :class="issueErrors.invoice_number ? 'input-error' : ''"
              required
            />
            <p v-if="issueErrors.invoice_number" class="input-error-text mt-1">{{ issueErrors.invoice_number }}</p>
          </div>
          <div>
            <label class="input-label" for="invoice-issue-code">{{ t('admin.invoices.issue.invoiceCode') }}</label>
            <input id="invoice-issue-code" v-model="issueForm.invoice_code" type="text" class="input mt-1 w-full" />
          </div>
        </div>
        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label" for="invoice-issue-amount">{{ t('admin.invoices.issue.invoiceAmount') }} *</label>
            <input
              id="invoice-issue-amount"
              v-model="issueForm.invoice_amount"
              type="number"
              step="0.01"
              min="0"
              class="input mt-1 w-full"
              :class="issueErrors.invoice_amount ? 'input-error' : ''"
              required
            />
            <p v-if="issueErrors.invoice_amount" class="input-error-text mt-1">{{ issueErrors.invoice_amount }}</p>
          </div>
          <div>
            <label class="input-label" for="invoice-issue-date">{{ t('admin.invoices.issue.issuedAt') }}</label>
            <input id="invoice-issue-date" v-model="issueForm.issued_at" type="datetime-local" class="input mt-1 w-full" />
          </div>
        </div>
        <div>
          <label class="input-label" for="invoice-issue-remark">{{ t('admin.invoices.issue.adminRemark') }}</label>
          <textarea id="invoice-issue-remark" v-model="issueForm.admin_remark" rows="2" class="input mt-1 w-full" />
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeIssueDialog">
            {{ t('common.cancel') }}
          </button>
          <button
            type="submit"
            form="invoice-issue-form"
            class="btn btn-primary"
            :disabled="submitting"
          >
            {{ submitting ? t('common.processing') : t('admin.invoices.issue.submit') }}
          </button>
        </div>
      </template>
    </BaseDialog>
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { adminInvoiceAPI } from '@/api/invoice'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import { formatPaymentAmount } from '@/components/payment/currency'
import type { Column } from '@/components/common/types'
import type {
  InvoiceApplication,
  InvoiceApplicationStatus,
  IssueInvoiceApplicationRequest,
  RejectInvoiceApplicationRequest
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import TablePageLayout from '@/components/layout/TablePageLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'
import Select from '@/components/common/Select.vue'

const { t } = useI18n()
const appStore = useAppStore()

const loading = ref(false)
const submitting = ref(false)
const applications = ref<InvoiceApplication[]>([])
const keyword = ref('')
const filters = reactive({ status: '' })
const pagination = reactive({ page: 1, page_size: 20, total: 0 })

const columns = computed<Column[]>(() => [
  { key: 'user', label: t('admin.invoices.columns.user') },
  { key: 'title', label: t('admin.invoices.columns.title') },
  { key: 'total_amount', label: t('admin.invoices.columns.amount') },
  { key: 'status', label: t('admin.invoices.columns.status') },
  { key: 'created_at', label: t('admin.invoices.columns.createdAt') },
  { key: 'actions', label: t('admin.invoices.columns.actions') }
])

const statusFilterOptions = computed(() => [
  { value: '', label: t('admin.invoices.allStatuses') },
  { value: 'PENDING', label: t('admin.invoices.status.pending') },
  { value: 'APPROVED', label: t('admin.invoices.status.approved') },
  { value: 'ISSUED', label: t('admin.invoices.status.issued') },
  { value: 'REJECTED', label: t('admin.invoices.status.rejected') },
  { value: 'CANCELLED', label: t('admin.invoices.status.cancelled') }
])

async function loadApplications() {
  loading.value = true
  try {
    const res = await adminInvoiceAPI.getApplications({
      page: pagination.page,
      page_size: pagination.page_size,
      status: filters.status || undefined,
      keyword: keyword.value || undefined
    })
    applications.value = res.data.items || []
    pagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    loading.value = false
  }
}

function handlePageChange(page: number) {
  pagination.page = page
  loadApplications()
}

function handlePageSizeChange(size: number) {
  pagination.page_size = size
  pagination.page = 1
  loadApplications()
}

function handleStatusChange() {
  pagination.page = 1
  loadApplications()
}

let searchDebounceTimer: number | null = null
function handleSearch() {
  if (searchDebounceTimer) window.clearTimeout(searchDebounceTimer)
  searchDebounceTimer = window.setTimeout(() => {
    pagination.page = 1
    loadApplications()
  }, 300)
}

// ==================== Detail ====================

const showDetailDialog = ref(false)
const detailApplication = ref<InvoiceApplication | null>(null)

async function openDetail(application: InvoiceApplication) {
  detailApplication.value = application
  showDetailDialog.value = true
  try {
    const res = await adminInvoiceAPI.getApplication(application.id)
    if (res.data) detailApplication.value = res.data
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  }
}

function closeDetailDialog() {
  showDetailDialog.value = false
  detailApplication.value = null
}

// ==================== Approve ====================

const approveTarget = ref<InvoiceApplication | null>(null)

async function confirmApprove() {
  const target = approveTarget.value
  approveTarget.value = null
  if (!target) return
  submitting.value = true
  try {
    await adminInvoiceAPI.approve(target.id)
    appStore.showSuccess(t('admin.invoices.approve.success'))
    await loadApplications()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    submitting.value = false
  }
}

// ==================== Reject ====================

const showRejectDialog = ref(false)
const rejectTarget = ref<InvoiceApplication | null>(null)
const rejectForm = reactive({ reason: '', remark: '' })
const rejectErrors = reactive({ reason: '' })

function openRejectDialog(application: InvoiceApplication) {
  rejectTarget.value = application
  rejectForm.reason = ''
  rejectForm.remark = ''
  rejectErrors.reason = ''
  showRejectDialog.value = true
}

function closeRejectDialog() {
  showRejectDialog.value = false
  rejectTarget.value = null
}

async function submitReject() {
  if (!rejectTarget.value) return
  if (!rejectForm.reason.trim()) {
    rejectErrors.reason = t('admin.invoices.reject.reasonRequired')
    return
  }
  const payload: RejectInvoiceApplicationRequest = {
    reason: rejectForm.reason.trim(),
    remark: rejectForm.remark.trim()
  }
  submitting.value = true
  try {
    await adminInvoiceAPI.reject(rejectTarget.value.id, payload)
    appStore.showSuccess(t('admin.invoices.reject.success'))
    closeRejectDialog()
    await loadApplications()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    submitting.value = false
  }
}

// ==================== Issue ====================

const showIssueDialog = ref(false)
const issueTarget = ref<InvoiceApplication | null>(null)
const issueForm = reactive({
  invoice_number: '',
  invoice_code: '',
  invoice_amount: '' as string | number,
  issued_at: '',
  admin_remark: ''
})
const issueErrors = reactive({ invoice_number: '', invoice_amount: '' })

function openIssueDialog(application: InvoiceApplication) {
  issueTarget.value = application
  issueForm.invoice_number = ''
  issueForm.invoice_code = ''
  issueForm.invoice_amount = application.total_amount
  issueForm.issued_at = ''
  issueForm.admin_remark = ''
  issueErrors.invoice_number = ''
  issueErrors.invoice_amount = ''
  showIssueDialog.value = true
}

function closeIssueDialog() {
  showIssueDialog.value = false
  issueTarget.value = null
}

async function submitIssue() {
  if (!issueTarget.value) return
  issueErrors.invoice_number = ''
  issueErrors.invoice_amount = ''

  if (!issueForm.invoice_number.trim()) {
    issueErrors.invoice_number = t('admin.invoices.issue.invoiceNumberRequired')
  }
  const amount = Number(issueForm.invoice_amount)
  if (!Number.isFinite(amount) || amount <= 0) {
    issueErrors.invoice_amount = t('admin.invoices.issue.invoiceAmountRequired')
  }
  if (issueErrors.invoice_number || issueErrors.invoice_amount) return

  const payload: IssueInvoiceApplicationRequest = {
    invoice_number: issueForm.invoice_number.trim(),
    invoice_code: issueForm.invoice_code.trim(),
    invoice_amount: amount,
    issued_at: issueForm.issued_at ? new Date(issueForm.issued_at).toISOString() : null,
    admin_remark: issueForm.admin_remark.trim()
  }

  submitting.value = true
  try {
    await adminInvoiceAPI.issue(issueTarget.value.id, payload)
    appStore.showSuccess(t('admin.invoices.issue.success'))
    closeIssueDialog()
    await loadApplications()
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    submitting.value = false
  }
}

// ==================== Display helpers ====================

function formatAmount(amount: number | null | undefined, currency?: string | null): string {
  return formatPaymentAmount(Number(amount ?? 0), currency ?? 'CNY')
}

function statusLabel(status: string): string {
  switch (status) {
    case 'PENDING': return t('admin.invoices.status.pending')
    case 'APPROVED': return t('admin.invoices.status.approved')
    case 'ISSUED': return t('admin.invoices.status.issued')
    case 'REJECTED': return t('admin.invoices.status.rejected')
    case 'CANCELLED': return t('admin.invoices.status.cancelled')
    default: return status
  }
}

function statusBadgeClass(status: string): string {
  switch (status as InvoiceApplicationStatus) {
    case 'PENDING': return 'badge-warning'
    case 'APPROVED': return 'badge-primary'
    case 'ISSUED': return 'badge-success'
    case 'REJECTED': return 'badge-danger'
    case 'CANCELLED': return 'badge-gray'
    default: return 'badge-gray'
  }
}

function invoiceTypeLabel(invoiceType: string): string {
  switch (invoiceType) {
    case 'electron_normal': return t('admin.invoices.invoiceType.electron_normal')
    case 'electron_special': return t('admin.invoices.invoiceType.electron_special')
    default: return invoiceType
  }
}

function titleTypeLabel(titleType: string): string {
  switch (titleType) {
    case 'personal': return t('admin.invoices.titleType.personal')
    case 'company': return t('admin.invoices.titleType.company')
    default: return titleType
  }
}

onMounted(() => loadApplications())
</script>
