<template>
  <AppLayout>
    <div class="space-y-4">
      <!-- Self-service invoicing guide -->
      <div class="card p-4">
        <h2 class="text-base font-semibold text-gray-900 dark:text-white">
          {{ t('invoice.guide.title') }}
        </h2>
        <p class="mt-1 text-sm text-gray-600 dark:text-gray-300">
          {{ t('invoice.guide.subtitle') }}
        </p>
        <div class="mt-2 flex flex-wrap items-center gap-x-5 gap-y-1">
          <p class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
            <Icon name="clock" class="h-3.5 w-3.5 shrink-0" />
            <span>{{ t('invoice.guide.tipSchedule') }}</span>
          </p>
          <p class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
            <Icon name="infoCircle" class="h-3.5 w-3.5 shrink-0" />
            <span>{{ t('invoice.guide.tipMonthEnd') }}</span>
          </p>
          <p class="flex items-center gap-1.5 text-xs text-gray-500 dark:text-gray-400">
            <Icon name="document" class="h-3.5 w-3.5 shrink-0" />
            <span>{{ t('invoice.guide.tipContent') }}</span>
          </p>
        </div>
      </div>

      <!-- Tabs -->
      <div class="card p-4">
        <div class="flex gap-2 border-b border-gray-200 dark:border-dark-600">
          <button
            v-for="tab in tabs"
            :key="tab.value"
            type="button"
            class="-mb-px border-b-2 px-4 py-2 text-sm font-medium transition-colors"
            :class="activeTab === tab.value
              ? 'border-primary-500 text-primary-600 dark:text-primary-400'
              : 'border-transparent text-gray-500 hover:border-gray-300 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200'"
            @click="activeTab = tab.value"
          >
            {{ tab.label }}
          </button>
        </div>
      </div>

      <!-- Eligible orders tab -->
      <template v-if="activeTab === 'eligible'">
        <!-- Selection summary -->
        <div class="card p-4">
          <div class="flex flex-wrap items-center justify-between gap-3">
            <div class="flex flex-wrap items-center gap-4 text-sm">
              <span class="text-gray-600 dark:text-gray-300">
                {{ t('invoice.selection.count', { count: selectedOrderIds.length }) }}
              </span>
              <span class="text-gray-600 dark:text-gray-300">
                {{ t('invoice.selection.amount') }}:
                <span class="font-semibold text-gray-900 dark:text-white">
                  {{ formatAmount(selectedTotal, displayCurrency) }}
                </span>
              </span>
              <span class="text-gray-600 dark:text-gray-300">
                {{ t('invoice.selection.minAmount') }}:
                <span class="font-semibold text-gray-900 dark:text-white">
                  {{ formatAmount(MIN_INVOICE_AMOUNT, displayCurrency) }}
                </span>
              </span>
            </div>
            <div class="flex items-center gap-3">
              <span
                v-if="selectedOrderIds.length > 0 && !meetsMinimum"
                class="text-sm font-medium text-red-600 dark:text-red-400"
              >
                {{ t('invoice.selection.needMore', { amount: formatAmount(minimumShortfall, displayCurrency) }) }}
              </span>
              <button
                class="btn btn-primary"
                :disabled="!canSubmit"
                @click="openApplyDialog"
              >
                {{ t('invoice.selection.apply') }}
              </button>
            </div>
          </div>
        </div>

        <!-- Eligible orders table -->
        <DataTable
          :columns="eligibleColumns"
          :data="eligibleOrders"
          :loading="eligibleLoading"
          selectable
          row-key="order_id"
          :selected-keys="selectedOrderIds"
          :selection-label="(row: InvoiceableOrder) => row.order_no"
          @update:selected-keys="handleOrderSelectionChange"
        >
          <template #cell-order_no="{ value }">
            <span class="font-mono text-xs text-gray-600 dark:text-gray-300">{{ value }}</span>
          </template>
          <template #cell-paid_at="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(value) }}</span>
          </template>
          <template #cell-pay_amount="{ value, row }">
            <span class="text-sm font-semibold text-gray-900 dark:text-white">
              {{ formatAmount(value, row.currency) }}
            </span>
          </template>
          <template #cell-total_amount="{ value, row }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ formatAmount(value, row.currency) }}</span>
          </template>
          <template #empty>
            <EmptyState :title="t('invoice.eligible.empty')" />
          </template>
        </DataTable>

        <Pagination
          v-if="eligiblePagination.total > 0"
          :page="eligiblePagination.page"
          :total="eligiblePagination.total"
          :page-size="eligiblePagination.page_size"
          @update:page="handleEligiblePageChange"
          @update:pageSize="handleEligiblePageSizeChange"
        />
      </template>

      <!-- My applications tab -->
      <template v-else>
        <DataTable
          :columns="applicationColumns"
          :data="applications"
          :loading="applicationsLoading"
        >
          <template #cell-id="{ value }">
            <span class="font-mono text-xs text-gray-500 dark:text-gray-400">#{{ value }}</span>
          </template>
          <template #cell-invoice_type="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ invoiceTypeLabel(value) }}</span>
          </template>
          <template #cell-title="{ value }">
            <span class="text-sm text-gray-900 dark:text-white">{{ value }}</span>
          </template>
          <template #cell-total_amount="{ value }">
            <span class="text-sm font-semibold text-gray-900 dark:text-white">{{ formatAmount(value, 'CNY') }}</span>
          </template>
          <template #cell-status="{ value }">
            <span :class="['badge', statusBadgeClass(value)]">{{ statusLabel(value) }}</span>
          </template>
          <template #cell-created_at="{ value }">
            <span class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(value) }}</span>
          </template>
          <template #cell-actions="{ row }">
            <div class="flex items-center gap-2">
              <button
                class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-gray-600 hover:bg-gray-100 dark:text-gray-400 dark:hover:bg-dark-600"
                @click="openApplicationDetail(row)"
              >
                <Icon name="eye" size="sm" />
                {{ t('invoice.applications.viewDetail') }}
              </button>
              <button
                v-if="row.status === 'PENDING'"
                class="inline-flex items-center gap-1 rounded-md px-2 py-1 text-xs font-medium text-yellow-600 hover:bg-yellow-50 dark:text-yellow-400 dark:hover:bg-yellow-900/20"
                @click="cancelTargetId = row.id"
              >
                <Icon name="x" size="sm" />
                {{ t('invoice.applications.cancel') }}
              </button>
            </div>
          </template>
          <template #empty>
            <EmptyState :title="t('invoice.applications.empty')" />
          </template>
        </DataTable>

        <Pagination
          v-if="applicationPagination.total > 0"
          :page="applicationPagination.page"
          :total="applicationPagination.total"
          :page-size="applicationPagination.page_size"
          @update:page="handleApplicationPageChange"
          @update:pageSize="handleApplicationPageSizeChange"
        />
      </template>
    </div>

    <!-- Apply for invoice dialog -->
    <BaseDialog
      :show="showApplyDialog"
      :title="t('invoice.form.title')"
      width="wide"
      @close="closeApplyDialog"
    >
      <form id="invoice-application-form" class="space-y-4" @submit.prevent="submitApplication">
        <div class="rounded-lg border border-blue-200 bg-blue-50 p-3 text-sm text-blue-800 dark:border-blue-800/50 dark:bg-blue-900/20 dark:text-blue-300">
          <p class="font-medium">{{ t('invoice.form.noticeTitle') }}</p>
          <ul class="mt-1 list-disc space-y-0.5 pl-5">
            <li>{{ t('invoice.form.noticeTax') }}</li>
            <li>{{ t('invoice.form.noticeEmail') }}</li>
          </ul>
        </div>

        <div>
          <label class="input-label" for="invoice-title">{{ t('invoice.form.invoiceTitle') }} *</label>
          <input
            id="invoice-title"
            v-model="form.title"
            type="text"
            class="input mt-1 w-full"
            :class="errors.title ? 'input-error' : ''"
            required
          />
          <p v-if="errors.title" class="input-error-text mt-1">{{ errors.title }}</p>
        </div>

        <div>
          <label class="input-label" for="invoice-tax-number">
            {{ t('invoice.form.taxNumber') }} *
          </label>
          <input
            id="invoice-tax-number"
            v-model="form.tax_number"
            type="text"
            class="input mt-1 w-full"
            :class="errors.tax_number ? 'input-error' : ''"
          />
          <p v-if="errors.tax_number" class="input-error-text mt-1">{{ errors.tax_number }}</p>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label" for="invoice-registered-address">{{ t('invoice.form.registeredAddress') }}</label>
            <input
              id="invoice-registered-address"
              v-model="form.registered_address"
              type="text"
              class="input mt-1 w-full"
            />
          </div>
          <div>
            <label class="input-label" for="invoice-registered-phone">{{ t('invoice.form.registeredPhone') }}</label>
            <input
              id="invoice-registered-phone"
              v-model="form.registered_phone"
              type="text"
              class="input mt-1 w-full"
            />
          </div>
        </div>

        <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <label class="input-label" for="invoice-bank-name">
              {{ t('invoice.form.bankName') }}
            </label>
            <input
              id="invoice-bank-name"
              v-model="form.bank_name"
              type="text"
              class="input mt-1 w-full"
            />
          </div>
          <div>
            <label class="input-label" for="invoice-bank-account">
              {{ t('invoice.form.bankAccount') }}
            </label>
            <input
              id="invoice-bank-account"
              v-model="form.bank_account"
              type="text"
              class="input mt-1 w-full"
            />
          </div>
        </div>

        <div>
          <label class="input-label" for="invoice-email">{{ t('invoice.form.email') }} *</label>
          <input
            id="invoice-email"
            v-model="form.email"
            type="email"
            class="input mt-1 w-full"
            :class="errors.email ? 'input-error' : ''"
            required
          />
          <p v-if="errors.email" class="input-error-text mt-1">{{ errors.email }}</p>
        </div>

        <div>
          <label class="input-label" for="invoice-remark">{{ t('invoice.form.remark') }}</label>
          <textarea id="invoice-remark" v-model="form.remark" rows="3" class="input mt-1 w-full" />
        </div>
      </form>
      <template #footer>
        <div class="flex justify-end gap-3">
          <button type="button" class="btn btn-secondary" @click="closeApplyDialog">
            {{ t('invoice.form.cancel') }}
          </button>
          <button
            type="submit"
            form="invoice-application-form"
            class="btn btn-primary"
            :disabled="submitting"
          >
            {{ submitting ? t('common.processing') : t('invoice.form.submit') }}
          </button>
        </div>
      </template>
    </BaseDialog>

    <!-- Application detail dialog -->
    <BaseDialog
      :show="!!detailApplication"
      :title="t('invoice.applications.viewDetail')"
      width="wide"
      @close="detailApplication = null"
    >
      <div v-if="detailApplication" class="space-y-4">
        <div class="grid grid-cols-2 gap-4">
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.id') }}</p>
            <p class="font-mono text-sm font-medium text-gray-900 dark:text-white">#{{ detailApplication.id }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.status') }}</p>
            <span :class="['badge', statusBadgeClass(detailApplication.status)]">
              {{ statusLabel(detailApplication.status) }}
            </span>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.invoiceType') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ invoiceTypeLabel(detailApplication.invoice_type) }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.title') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ detailApplication.title }}</p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.amount') }}</p>
            <p class="text-sm font-medium text-gray-900 dark:text-white">
              {{ formatAmount(detailApplication.total_amount, 'CNY') }}
            </p>
          </div>
          <div>
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.createdAt') }}</p>
            <p class="text-sm text-gray-700 dark:text-gray-300">{{ formatDateTime(detailApplication.created_at) }}</p>
          </div>
          <div v-if="detailApplication.invoice_number">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.invoiceNumber') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">{{ detailApplication.invoice_number }}</p>
          </div>
          <div v-if="detailApplication.invoice_amount != null">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.invoiceAmount') }}</p>
            <p class="text-sm text-gray-900 dark:text-white">
              {{ formatAmount(detailApplication.invoice_amount, 'CNY') }}
            </p>
          </div>
          <div v-if="detailApplication.rejected_reason" class="col-span-2">
            <p class="text-xs text-gray-500 dark:text-gray-400">{{ t('invoice.applications.rejectedReason') }}</p>
            <p class="text-sm text-red-600 dark:text-red-400">{{ detailApplication.rejected_reason }}</p>
          </div>
        </div>
      </div>
    </BaseDialog>

    <!-- Cancel confirmation -->
    <ConfirmDialog
      :show="!!cancelTargetId"
      :title="t('invoice.applications.cancel')"
      :message="t('invoice.applications.cancelConfirm')"
      :confirm-text="t('invoice.applications.cancel')"
      :cancel-text="t('common.cancel')"
      danger
      @confirm="confirmCancel"
      @cancel="cancelTargetId = null"
    />
  </AppLayout>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { invoiceAPI, MIN_INVOICE_AMOUNT } from '@/api/invoice'
import { extractI18nErrorMessage } from '@/utils/apiError'
import { formatDateTime } from '@/utils/format'
import { formatPaymentAmount } from '@/components/payment/currency'
import type { Column } from '@/components/common/types'
import type {
  CreateInvoiceApplicationRequest,
  InvoiceApplication,
  InvoiceApplicationStatus,
  InvoiceableOrder
} from '@/types'
import AppLayout from '@/components/layout/AppLayout.vue'
import BaseDialog from '@/components/common/BaseDialog.vue'
import ConfirmDialog from '@/components/common/ConfirmDialog.vue'
import DataTable from '@/components/common/DataTable.vue'
import EmptyState from '@/components/common/EmptyState.vue'
import Icon from '@/components/icons/Icon.vue'
import Pagination from '@/components/common/Pagination.vue'

const { t } = useI18n()
const appStore = useAppStore()

type TabValue = 'eligible' | 'applications'

const tabs = computed(() => [
  { value: 'eligible' as TabValue, label: t('invoice.tabEligible') },
  { value: 'applications' as TabValue, label: t('invoice.tabMyApplications') }
])

const activeTab = ref<TabValue>('eligible')

// ==================== Eligible orders ====================

const eligibleLoading = ref(false)
const eligibleOrders = ref<InvoiceableOrder[]>([])
const eligiblePagination = reactive({ page: 1, page_size: 20, total: 0 })
/** order_id -> pay_amount, accumulated across pages so cross-page selection keeps its amount */
const orderAmounts = reactive<Record<string, number>>({})
const orderCurrencies = reactive<Record<string, string>>({})
const selectedOrderIds = ref<number[]>([])

const eligibleColumns = computed<Column[]>(() => [
  { key: 'order_no', label: t('invoice.eligible.orderNo') },
  { key: 'paid_at', label: t('invoice.eligible.paidAt') },
  { key: 'pay_amount', label: t('invoice.eligible.payAmount') },
  { key: 'total_amount', label: t('invoice.eligible.creditedAmount') }
])

const selectedTotal = computed(() =>
  selectedOrderIds.value.reduce((sum, id) => sum + (orderAmounts[String(id)] ?? 0), 0)
)

const selectedCurrency = computed(() => {
  const first = selectedOrderIds.value[0]
  return first != null ? (orderCurrencies[String(first)] ?? 'CNY') : 'CNY'
})

/**
 * Currency used for every amount shown in the selection bar. The selected total,
 * the minimum threshold and the shortfall all live in the same numeric space
 * (the backend compares the raw pay_amount sum against its minimum), so they
 * must be rendered with one unit to stay comparable.
 */
const displayCurrency = computed(() => selectedCurrency.value)

const meetsMinimum = computed(() => selectedTotal.value >= MIN_INVOICE_AMOUNT)
const minimumShortfall = computed(() => Math.max(0, MIN_INVOICE_AMOUNT - selectedTotal.value))
const canSubmit = computed(
  () => selectedOrderIds.value.length > 0 && meetsMinimum.value
)

async function fetchEligibleOrders() {
  eligibleLoading.value = true
  try {
    const res = await invoiceAPI.getEligibleOrders({
      page: eligiblePagination.page,
      page_size: eligiblePagination.page_size
    })
    eligibleOrders.value = res.data.items || []
    eligiblePagination.total = res.data.total || 0
    for (const order of eligibleOrders.value) {
      orderAmounts[String(order.order_id)] = order.pay_amount
      orderCurrencies[String(order.order_id)] = order.currency
    }
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    eligibleLoading.value = false
  }
}

function handleEligiblePageChange(page: number) {
  eligiblePagination.page = page
  fetchEligibleOrders()
}

function handleEligiblePageSizeChange(size: number) {
  eligiblePagination.page_size = size
  eligiblePagination.page = 1
  fetchEligibleOrders()
}

function handleOrderSelectionChange(ids: Array<string | number>) {
  selectedOrderIds.value = [...new Set(ids.map(Number))].filter((id) => !Number.isNaN(id))
}

// ==================== My applications ====================

const applicationsLoading = ref(false)
const applications = ref<InvoiceApplication[]>([])
const applicationPagination = reactive({ page: 1, page_size: 20, total: 0 })
const detailApplication = ref<InvoiceApplication | null>(null)
const cancelTargetId = ref<number | null>(null)

const applicationColumns = computed<Column[]>(() => [
  { key: 'id', label: t('invoice.applications.id') },
  { key: 'invoice_type', label: t('invoice.applications.invoiceType') },
  { key: 'title', label: t('invoice.applications.title') },
  { key: 'total_amount', label: t('invoice.applications.amount') },
  { key: 'status', label: t('invoice.applications.status') },
  { key: 'created_at', label: t('invoice.applications.createdAt') },
  { key: 'actions', label: t('common.actions') }
])

async function fetchApplications() {
  applicationsLoading.value = true
  try {
    const res = await invoiceAPI.getMyApplications({
      page: applicationPagination.page,
      page_size: applicationPagination.page_size
    })
    applications.value = res.data.items || []
    applicationPagination.total = res.data.total || 0
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  } finally {
    applicationsLoading.value = false
  }
}

function handleApplicationPageChange(page: number) {
  applicationPagination.page = page
  fetchApplications()
}

function handleApplicationPageSizeChange(size: number) {
  applicationPagination.page_size = size
  applicationPagination.page = 1
  fetchApplications()
}

function openApplicationDetail(application: InvoiceApplication) {
  detailApplication.value = application
}

async function confirmCancel() {
  if (!cancelTargetId.value) return
  const id = cancelTargetId.value
  cancelTargetId.value = null
  try {
    await invoiceAPI.cancelApplication(id)
    appStore.showSuccess(t('invoice.applications.cancelSuccess'))
    await Promise.all([fetchApplications(), fetchEligibleOrders()])
  } catch (err: unknown) {
    appStore.showError(extractI18nErrorMessage(err, t, 'invoice.errors', t('common.error')))
  }
}

// ==================== Apply form ====================

const showApplyDialog = ref(false)
const submitting = ref(false)

const form = reactive({
  title: '',
  tax_number: '',
  registered_address: '',
  registered_phone: '',
  bank_name: '',
  bank_account: '',
  email: '',
  remark: ''
})

const errors = reactive({
  title: '',
  tax_number: '',
  email: ''
})

function openApplyDialog() {
  if (!canSubmit.value) {
    appStore.showError(t('invoice.selection.noneSelected'))
    return
  }
  errors.title = ''
  errors.tax_number = ''
  errors.email = ''
  showApplyDialog.value = true
}

function closeApplyDialog() {
  showApplyDialog.value = false
}

function validateForm(): boolean {
  errors.title = form.title.trim() ? '' : t('invoice.form.titleRequired')
  errors.tax_number = form.tax_number.trim() ? '' : t('invoice.form.taxNumberRequired')
  errors.email = ''

  if (!form.email.trim()) {
    errors.email = t('invoice.form.emailRequired')
  } else if (!/^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(form.email.trim())) {
    errors.email = t('invoice.form.emailInvalid')
  }

  return !errors.title && !errors.tax_number && !errors.email
}

async function submitApplication() {
  if (!validateForm()) return
  if (!meetsMinimum.value) {
    appStore.showError(t('invoice.selection.needMore', { amount: formatAmount(minimumShortfall.value, displayCurrency.value) }))
    return
  }

  const payload: CreateInvoiceApplicationRequest = {
    order_ids: selectedOrderIds.value,
    invoice_type: 'electron_normal',
    title_type: 'company',
    title: form.title.trim(),
    tax_number: form.tax_number.trim(),
    registered_address: form.registered_address.trim(),
    registered_phone: form.registered_phone.trim(),
    bank_name: form.bank_name.trim(),
    bank_account: form.bank_account.trim(),
    email: form.email.trim(),
    remark: form.remark.trim()
  }

  submitting.value = true
  try {
    await invoiceAPI.createApplication(payload)
    appStore.showSuccess(t('invoice.applications.submitSuccess'))
    closeApplyDialog()
    selectedOrderIds.value = []
    form.title = ''
    form.tax_number = ''
    form.registered_address = ''
    form.registered_phone = ''
    form.bank_name = ''
    form.bank_account = ''
    form.remark = ''
    activeTab.value = 'applications'
    applicationPagination.page = 1
    await Promise.all([fetchApplications(), fetchEligibleOrders()])
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
    case 'PENDING': return t('invoice.status.pending')
    case 'APPROVED': return t('invoice.status.approved')
    case 'ISSUED': return t('invoice.status.issued')
    case 'REJECTED': return t('invoice.status.rejected')
    case 'CANCELLED': return t('invoice.status.cancelled')
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
    case 'electron_normal': return t('invoice.invoiceType.electron_normal')
    case 'electron_special': return t('invoice.invoiceType.electron_special')
    default: return invoiceType
  }
}

onMounted(() => {
  fetchEligibleOrders()
  fetchApplications()
})
</script>
