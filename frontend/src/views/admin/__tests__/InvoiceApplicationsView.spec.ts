import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'

import InvoiceApplicationsView from '../InvoiceApplicationsView.vue'
import DataTable from '@/components/common/DataTable.vue'
import Select from '@/components/common/Select.vue'
import Pagination from '@/components/common/Pagination.vue'

const api = vi.hoisted(() => ({
  getApplications: vi.fn(),
  getApplication: vi.fn(),
  approve: vi.fn(),
  reject: vi.fn(),
  issue: vi.fn()
}))

const store = vi.hoisted(() => ({
  showSuccess: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/invoice', () => ({ adminInvoiceAPI: api }))
vi.mock('@/stores/app', () => ({
  useAppStore: () => store
}))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key })
}))

enableAutoUnmount(afterEach)

const pendingApplication = {
  id: 11,
  user_id: 3,
  user_email: 'buyer@example.com',
  user_name: 'Buyer',
  status: 'PENDING',
  invoice_type: 'electron_normal',
  title_type: 'company',
  title: 'Acme Ltd',
  tax_number: '91310000XXXXXXXXXX',
  registered_address: 'Somewhere',
  registered_phone: '021-00000000',
  bank_name: '',
  bank_account: '',
  email: 'buyer@example.com',
  remark: '',
  total_amount: 500,
  invoice_amount: null,
  invoice_number: '',
  invoice_code: '',
  issued_at: null,
  reviewed_by: null,
  reviewed_at: null,
  rejected_reason: '',
  admin_remark: '',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z',
  items: [
    {
      id: 1,
      order_id: 7,
      order_no: 'sub2_order_1',
      amount: 500,
      currency: 'CNY',
      paid_at: '2026-01-01T00:00:00Z',
      released_at: null
    }
  ]
}

function mountView() {
  return mount(InvoiceApplicationsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        TablePageLayout: {
          template: '<div><slot name="filters" /><slot name="table" /><slot name="pagination" /></div>'
        },
        DataTable: true,
        Pagination: true,
        Select: true,
        BaseDialog: true,
        ConfirmDialog: true,
        EmptyState: true,
        Icon: true,
        Teleport: true
      }
    }
  })
}

describe('admin invoice applications list', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.getApplications.mockResolvedValue({
      data: {
        items: [pendingApplication],
        total: 1,
        page: 1,
        page_size: 20,
        pages: 1
      }
    })
    api.getApplication.mockResolvedValue({ data: pendingApplication })
    api.approve.mockResolvedValue({ data: { message: 'ok' } })
  })

  it('loads the first page of invoice applications on mount', async () => {
    const wrapper = mountView()
    await flushPromises()

    expect(api.getApplications).toHaveBeenCalledWith({
      page: 1,
      page_size: 20,
      status: undefined,
      keyword: undefined
    })
    expect(wrapper.getComponent(DataTable).props('data')).toHaveLength(1)
  })

  it('requests the next page when pagination changes', async () => {
    const wrapper = mountView()
    await flushPromises()

    wrapper.getComponent(Pagination).vm.$emit('update:page', 2)
    await flushPromises()

    expect(api.getApplications).toHaveBeenLastCalledWith({
      page: 2,
      page_size: 20,
      status: undefined,
      keyword: undefined
    })
  })

  it('filters by status', async () => {
    const wrapper = mountView()
    await flushPromises()

    const select = wrapper.getComponent(Select)
    await select.vm.$emit('update:modelValue', 'PENDING')
    await select.vm.$emit('change', 'PENDING')
    await flushPromises()

    expect(api.getApplications).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      status: 'PENDING',
      keyword: undefined
    })
  })

  it('searches by keyword', async () => {
    vi.useFakeTimers()
    const wrapper = mountView()
    await vi.advanceTimersByTimeAsync(400)

    const input = wrapper.find('input[type="text"]')
    await input.setValue('Acme')
    await vi.advanceTimersByTimeAsync(400)
    await flushPromises()

    expect(api.getApplications).toHaveBeenLastCalledWith({
      page: 1,
      page_size: 20,
      status: undefined,
      keyword: 'Acme'
    })
    vi.useRealTimers()
  })
})
