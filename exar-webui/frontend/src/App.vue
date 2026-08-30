<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { api } from './api'

const persons = ref([])
const shops = ref([])
const invoices = ref([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const showForm = ref(false)

const filters = reactive({
  person_id: '',
  from_date: '',
  to_date: '',
})

const form = reactive({
  person_id: 1,
  shop_id: '',
  newShopName: '',
  useNewShop: false,
  date: new Date().toISOString().slice(0, 10),
  items: [{ description: '', amount: '', quantity: 1 }],
})

const grandTotal = computed(() =>
  (invoices.value ?? []).reduce((sum, inv) => sum + inv.total, 0)
)

const lineTotal = computed(() =>
  form.items.reduce((sum, item) => {
    const amount = parseFloat(item.amount) || 0
    const qty = parseFloat(item.quantity) || 1
    return sum + amount * qty
  }, 0)
)

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const [p, s, inv] = await Promise.all([
      api.getPersons(),
      api.getShops(),
      api.getInvoices({
        person_id: filters.person_id || undefined,
        from_date: filters.from_date || undefined,
        to_date: filters.to_date || undefined,
      }),
    ])
    persons.value = p ?? []
    shops.value = s ?? []
    invoices.value = inv ?? []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function addLineItem() {
  form.items.push({ description: '', amount: '', quantity: 1 })
}

function removeLineItem(index) {
  if (form.items.length > 1) form.items.splice(index, 1)
}

async function submitExpense() {
  saving.value = true
  error.value = ''
  try {
    let shopId = form.shop_id
    if (form.useNewShop) {
      const shop = await api.createShop(form.newShopName.trim())
      shopId = shop.id
      shops.value.push(shop)
    }
    if (!shopId) throw new Error('Select or create a shop')

    const payload = {
      person_id: Number(form.person_id),
      shop_id: Number(shopId),
      date: form.date,
      items: form.items.map((item) => ({
        description: item.description.trim(),
        amount: parseFloat(item.amount),
        quantity: parseFloat(item.quantity) || 1,
      })),
    }

    await api.createInvoice(payload)
    showForm.value = false
    form.shop_id = ''
    form.newShopName = ''
    form.useNewShop = false
    form.items = [{ description: '', amount: '', quantity: 1 }]
    await loadData()
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

async function deleteInvoice(id) {
  if (!confirm('Delete this expense?')) return
  try {
    await api.deleteInvoice(id)
    await loadData()
  } catch (e) {
    error.value = e.message
  }
}

function formatMoney(value) {
  return new Intl.NumberFormat(undefined, {
    style: 'currency',
    currency: 'USD',
  }).format(value)
}

function formatDate(dateStr) {
  return new Date(dateStr + 'T00:00:00').toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

onMounted(loadData)
</script>

<template>
  <div class="min-h-screen">
    <header class="bg-brand-700 text-white shadow">
      <div class="mx-auto flex max-w-5xl items-center justify-between px-4 py-5 sm:px-6">
        <div>
          <h1 class="text-2xl font-bold tracking-tight">exar</h1>
          <p class="mt-1 text-sm text-brand-100">Track who spent what, where, and when</p>
        </div>
        <button
          class="rounded-lg bg-white px-4 py-2 text-sm font-semibold text-brand-700 shadow hover:bg-brand-50"
          @click="showForm = !showForm"
        >
          {{ showForm ? 'Cancel' : '+ Add expense' }}
        </button>
      </div>
    </header>

    <main class="mx-auto max-w-5xl space-y-6 px-4 py-6 sm:px-6">
      <div
        v-if="error"
        class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-700"
      >
        {{ error }}
      </div>

      <section
        v-if="showForm"
        class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm"
      >
        <h2 class="text-lg font-semibold text-slate-800">New expense</h2>
        <form class="mt-4 space-y-4" @submit.prevent="submitExpense">
          <div class="grid gap-4 sm:grid-cols-2">
            <label class="block text-sm">
              <span class="font-medium text-slate-700">Person</span>
              <select
                v-model="form.person_id"
                class="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2"
                required
              >
                <option v-for="p in persons" :key="p.id" :value="p.id">
                  {{ p.name }}
                </option>
              </select>
            </label>

            <label class="block text-sm">
              <span class="font-medium text-slate-700">Date</span>
              <input
                v-model="form.date"
                type="date"
                class="mt-1 w-full rounded-lg border border-slate-300 px-3 py-2"
                required
              />
            </label>
          </div>

          <div class="space-y-2">
            <span class="text-sm font-medium text-slate-700">Shop</span>
            <div class="flex flex-wrap gap-3">
              <label class="inline-flex items-center gap-2 text-sm">
                <input v-model="form.useNewShop" type="radio" :value="false" />
                Existing
              </label>
              <label class="inline-flex items-center gap-2 text-sm">
                <input v-model="form.useNewShop" type="radio" :value="true" />
                New shop
              </label>
            </div>
            <select
              v-if="!form.useNewShop"
              v-model="form.shop_id"
              class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
              required
            >
              <option disabled value="">Select shop</option>
              <option v-for="s in shops" :key="s.id" :value="s.id">{{ s.name }}</option>
            </select>
            <input
              v-else
              v-model="form.newShopName"
              type="text"
              placeholder="Shop name"
              class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm"
              required
            />
          </div>

          <div>
            <div class="mb-2 flex items-center justify-between">
              <span class="text-sm font-medium text-slate-700">Line items</span>
              <button
                type="button"
                class="text-sm font-medium text-brand-600 hover:text-brand-700"
                @click="addLineItem"
              >
                + Add item
              </button>
            </div>
            <div class="space-y-2">
              <div
                v-for="(item, index) in form.items"
                :key="index"
                class="grid gap-2 rounded-lg bg-slate-50 p-3 sm:grid-cols-12"
              >
                <input
                  v-model="item.description"
                  type="text"
                  placeholder="Description"
                  class="rounded border border-slate-300 px-2 py-1.5 text-sm sm:col-span-5"
                  required
                />
                <input
                  v-model="item.amount"
                  type="number"
                  step="0.01"
                  min="0"
                  placeholder="Amount"
                  class="rounded border border-slate-300 px-2 py-1.5 text-sm sm:col-span-3"
                  required
                />
                <input
                  v-model="item.quantity"
                  type="number"
                  step="0.01"
                  min="0.01"
                  placeholder="Qty"
                  class="rounded border border-slate-300 px-2 py-1.5 text-sm sm:col-span-2"
                />
                <button
                  type="button"
                  class="text-sm text-red-600 hover:text-red-700 sm:col-span-2"
                  :disabled="form.items.length === 1"
                  @click="removeLineItem(index)"
                >
                  Remove
                </button>
              </div>
            </div>
            <p class="mt-2 text-right text-sm text-slate-600">
              Total: <span class="font-semibold">{{ formatMoney(lineTotal) }}</span>
            </p>
          </div>

          <button
            type="submit"
            :disabled="saving"
            class="w-full rounded-lg bg-brand-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-brand-700 disabled:opacity-60 sm:w-auto"
          >
            {{ saving ? 'Saving…' : 'Save expense' }}
          </button>
        </form>
      </section>

      <section class="rounded-xl border border-slate-200 bg-white p-5 shadow-sm">
        <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <h2 class="text-lg font-semibold text-slate-800">Expenses</h2>
            <p class="text-sm text-slate-500">
              {{ invoices.length }} records · {{ formatMoney(grandTotal) }} total
            </p>
          </div>
          <div class="grid gap-2 sm:grid-cols-3">
            <select
              v-model="filters.person_id"
              class="rounded-lg border border-slate-300 px-3 py-2 text-sm"
              @change="loadData"
            >
              <option value="">All people</option>
              <option v-for="p in persons" :key="p.id" :value="p.id">{{ p.name }}</option>
            </select>
            <input
              v-model="filters.from_date"
              type="date"
              class="rounded-lg border border-slate-300 px-3 py-2 text-sm"
              @change="loadData"
            />
            <input
              v-model="filters.to_date"
              type="date"
              class="rounded-lg border border-slate-300 px-3 py-2 text-sm"
              @change="loadData"
            />
          </div>
        </div>

        <div v-if="loading" class="py-12 text-center text-slate-500">Loading…</div>
        <div v-else-if="invoices.length === 0" class="py-12 text-center text-slate-500">
          No expenses yet. Add your first one above.
        </div>
        <ul v-else class="mt-4 divide-y divide-slate-100">
          <li
            v-for="inv in invoices"
            :key="inv.id"
            class="flex flex-col gap-3 py-4 sm:flex-row sm:items-start sm:justify-between"
          >
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span class="rounded-full bg-brand-100 px-2.5 py-0.5 text-xs font-medium text-brand-700">
                  {{ inv.person_name }}
                </span>
                <span class="text-sm font-medium text-slate-800">{{ inv.shop_name }}</span>
                <span class="text-sm text-slate-500">{{ formatDate(inv.date) }}</span>
              </div>
              <ul class="mt-2 space-y-1 text-sm text-slate-600">
                <li v-for="item in inv.items" :key="item.id">
                  {{ item.description }}
                  <span class="text-slate-400">×{{ item.quantity }}</span>
                  — {{ formatMoney(item.amount * item.quantity) }}
                </li>
              </ul>
            </div>
            <div class="flex items-center gap-3 sm:flex-col sm:items-end">
              <span class="text-lg font-semibold text-slate-900">{{ formatMoney(inv.total) }}</span>
              <button
                class="text-sm text-red-600 hover:text-red-700"
                @click="deleteInvoice(inv.id)"
              >
                Delete
              </button>
            </div>
          </li>
        </ul>
      </section>
    </main>
  </div>
</template>
