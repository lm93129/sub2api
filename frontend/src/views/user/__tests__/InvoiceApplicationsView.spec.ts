import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { enableAutoUnmount, flushPromises, mount } from '@vue/test-utils'

import InvoiceApplicationsView from '../InvoiceApplicationsView.vue'
import DataTable from '@/components/common/DataTable.vue'

const api = vi.hoisted(() => ({
  getEligibleOrders: vi.fn(),
  getMyApplications: vi.fn(),
  getApplication: vi.fn(),
  createApplication: vi.fn(),
  cancelApplication: vi.fn()
}))

const store = vi.hoisted(() => ({
  showSuccess: vi.fn(),
  showError: vi.fn()
}))

vi.mock('@/api/invoice', () => ({
  invoiceAPI: api,
  MIN_INVOICE_AMOUNT: 300
}))
vi.mock('@/stores/app', () => ({
  useAppStore: () => store
}))
vi.mock('vue-i18n', async (importOriginal) => ({
  ...(await importOriginal<typeof import('vue-i18n')>()),
  useI18n: () => ({ t: (key: string) => key })
}))

enableAutoUnmount(afterEach)

function eligibleOrder(orderId: number, payAmount: number) {
  return {
    order_id: orderId,
    order_no: `sub2_order_${orderId}`,
    pay_amount: payAmount,
    total_amount: payAmount,
    currency: 'CNY',
    paid_at: '2026-01-01T00:00:00Z',
    created_at: '2026-01-01T00:00:00Z'
  }
}

/** Renders the dialog body so the form can be driven from the test. */
const BaseDialogStub = {
  props: ['show', 'title', 'width'],
  template: '<div v-if="show" data-test="dialog"><slot /><slot name="footer" /></div>'
}

/** Native select stub so v-model on Select can be driven from the test. */
const SelectStub = {
  props: ['modelValue', 'options'],
  emits: ['update:modelValue', 'change'],
  setup(
    props: { options: Array<{ value: unknown; label: string }> },
    { emit }: { emit: (event: string, ...args: unknown[]) => void }
  ) {
    const onChange = (event: Event) => {
      const raw = (event.target as HTMLSelectElement).value
      const option = props.options.find((item) => String(item.value ?? '') === raw)
      const value = option ? option.value : raw
      emit('update:modelValue', value)
      emit('change', value, option ?? null)
    }
    return { onChange }
  },
  template: `
    <select :value="modelValue ?? ''" @change="onChange">
      <option v-for="option in options" :key="String(option.value ?? '')" :value="option.value ?? ''">
        {{ option.label }}
      </option>
    </select>
  `
}

function mountView() {
  return mount(InvoiceApplicationsView, {
    global: {
      stubs: {
        AppLayout: { template: '<div><slot /></div>' },
        DataTable: true,
        Pagination: true,
        Select: SelectStub,
        BaseDialog: BaseDialogStub,
        ConfirmDialog: true,
        EmptyState: true,
        Icon: true,
        Teleport: true
      }
    }
  })
}

function findApplyButton(wrapper: ReturnType<typeof mountView>) {
  return wrapper.findAll('button').find((button) => button.text().includes('invoice.selection.apply'))
}

describe('user invoice application', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    api.getEligibleOrders.mockResolvedValue({
      data: {
        items: [eligibleOrder(1, 200), eligibleOrder(2, 500)],
        total: 2,
        page: 1,
        page_size: 20,
        pages: 1
      }
    })
    api.getMyApplications.mockResolvedValue({
      data: { items: [], total: 0, page: 1, page_size: 20, pages: 1 }
    })
    api.createApplication.mockResolvedValue({ data: { id: 1 } })
  })

  it('loads eligible orders and my applications on mount', async () => {
    mountView()
    await flushPromises()

    expect(api.getEligibleOrders).toHaveBeenCalledWith({ page: 1, page_size: 20 })
    expect(api.getMyApplications).toHaveBeenCalledWith({ page: 1, page_size: 20 })
  })

  it('blocks the apply action when the selected total is below the minimum', async () => {
    const wrapper = mountView()
    await flushPromises()

    // 200 CNY is below the 300 CNY minimum.
    wrapper.getComponent(DataTable).vm.$emit('update:selectedKeys', [1])
    await flushPromises()

    const applyButton = findApplyButton(wrapper)
    expect(applyButton).toBeDefined()
    expect(applyButton!.attributes('disabled')).toBeDefined()
    expect(wrapper.text()).toContain('invoice.selection.needMore')
  })

  it('enables the apply action once the selected total reaches the minimum', async () => {
    const wrapper = mountView()
    await flushPromises()

    // 500 CNY reaches the 300 CNY minimum.
    wrapper.getComponent(DataTable).vm.$emit('update:selectedKeys', [2])
    await flushPromises()

    const applyButton = findApplyButton(wrapper)
    expect(applyButton!.attributes('disabled')).toBeUndefined()
    expect(wrapper.text()).not.toContain('invoice.selection.needMore')
  })

  it('requires a tax number for company titles and blocks submission', async () => {
    const wrapper = mountView()
    await flushPromises()

    wrapper.getComponent(DataTable).vm.$emit('update:selectedKeys', [2])
    await flushPromises()

    await findApplyButton(wrapper)!.trigger('click')
    await flushPromises()

    const selects = wrapper.findAll('select')
    await selects[1].setValue('company')
    await flushPromises()

    await wrapper.find('#invoice-title').setValue('Acme Ltd')
    await wrapper.find('#invoice-email').setValue('buyer@example.com')
    await wrapper.find('#invoice-application-form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('invoice.form.taxNumberRequired')
    expect(api.createApplication).not.toHaveBeenCalled()
  })

  it('requires bank details for electronic special VAT invoices', async () => {
    const wrapper = mountView()
    await flushPromises()

    wrapper.getComponent(DataTable).vm.$emit('update:selectedKeys', [2])
    await flushPromises()

    await findApplyButton(wrapper)!.trigger('click')
    await flushPromises()

    const selects = wrapper.findAll('select')
    await selects[0].setValue('electron_special')
    await flushPromises()

    await wrapper.find('#invoice-title').setValue('Acme Ltd')
    await wrapper.find('#invoice-email').setValue('buyer@example.com')
    await wrapper.find('#invoice-application-form').trigger('submit')
    await flushPromises()

    expect(wrapper.text()).toContain('invoice.form.bankInfoRequired')
    expect(api.createApplication).not.toHaveBeenCalled()
  })

  it('submits the application with the selected orders and the invoice form', async () => {
    const wrapper = mountView()
    await flushPromises()

    wrapper.getComponent(DataTable).vm.$emit('update:selectedKeys', [2])
    await flushPromises()

    await findApplyButton(wrapper)!.trigger('click')
    await flushPromises()

    await wrapper.find('#invoice-title').setValue('Acme Ltd')
    await wrapper.find('#invoice-email').setValue('buyer@example.com')
    await wrapper.find('#invoice-remark').setValue('please issue')
    await wrapper.find('#invoice-application-form').trigger('submit')
    await flushPromises()

    expect(api.createApplication).toHaveBeenCalledWith(
      expect.objectContaining({
        order_ids: [2],
        invoice_type: 'electron_normal',
        title_type: 'personal',
        title: 'Acme Ltd',
        email: 'buyer@example.com',
        remark: 'please issue'
      })
    )
  })
})
