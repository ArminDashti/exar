<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { jalaliToGregorian, toJalaliDisplay } from '../utils/dates'
import { formatMoney } from '../utils/money'
import JalaliDateInput from '../components/JalaliDateInput.vue'

const router = useRouter()
const persons = ref([])
const expenses = ref([])
const loading = ref(true)
const error = ref('')
const duplicateWarning = ref('')
const duplicateAcknowledged = ref(false)
const editingId = ref(null)
const saving = ref(false)

const filters = reactive({
  person_id: '',
  from_date: '',
  to_date: '',
})

const editForm = reactive({
  person_id: 1,
  shop_name: '',
  shop_id: null,
  date: '',
  name: '',
  amount: '',
  armin_share: 0.5,
  ramin_share: 0.5,
  shopSuggestions: [],
  showShopSuggestions: false,
})

let shopTimer = null

const grandTotal = computed(() =>
  (expenses.value ?? []).reduce((sum, e) => sum + e.amount, 0)
)

function parseWholeAmount(value) {
  if (value === '' || value === null || value === undefined) return NaN
  const text = String(value).trim()
  if (!/^\d+$/.test(text)) return NaN
  return Number(text)
}

function gregorianFilter() {
  const from = filters.from_date ? jalaliToGregorian(filters.from_date) : undefined
  const to = filters.to_date ? jalaliToGregorian(filters.to_date) : undefined
  if (filters.from_date && !from) throw new Error('Invalid from date')
  if (filters.to_date && !to) throw new Error('Invalid to date')
  return { from_date: from || undefined, to_date: to || undefined }
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const dates = gregorianFilter()
    const [p, list] = await Promise.all([
      api.getPersons(),
      api.getExpenses({
        person_id: filters.person_id || undefined,
        ...dates,
      }),
    ])
    persons.value = p ?? []
    expenses.value = list ?? []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function shareOf(exp, personId) {
  return exp.shares?.find((s) => s.person_id === personId)?.share ?? 0
}

function setEditShare(person, value) {
  const n = parseFloat(value)
  if (Number.isNaN(n)) {
    if (person === 'armin') editForm.armin_share = value
    else editForm.ramin_share = value
    return
  }
  const clamped = Math.min(1, Math.max(0, n))
  if (person === 'armin') {
    editForm.armin_share = clamped
    editForm.ramin_share = Math.round((1 - clamped) * 1000) / 1000
  } else {
    editForm.ramin_share = clamped
    editForm.armin_share = Math.round((1 - clamped) * 1000) / 1000
  }
}

function clearDuplicateWarning() {
  duplicateWarning.value = ''
  duplicateAcknowledged.value = false
}

function startEdit(exp) {
  editingId.value = exp.id
  editForm.person_id = exp.person_id
  editForm.shop_name = exp.shop_name
  editForm.shop_id = exp.shop_id
  editForm.date = toJalaliDisplay(exp.date)
  editForm.name = exp.name
  editForm.amount = String(Math.trunc(exp.amount))
  editForm.armin_share = shareOf(exp, 1)
  editForm.ramin_share = shareOf(exp, 2)
  editForm.shopSuggestions = []
  editForm.showShopSuggestions = false
  clearDuplicateWarning()
}

function cancelEdit() {
  editingId.value = null
  clearDuplicateWarning()
}

function onShopInput() {
  editForm.shop_id = null
  clearTimeout(shopTimer)
  const q = editForm.shop_name.trim()
  if ([...q].length < 3) {
    editForm.shopSuggestions = []
    editForm.showShopSuggestions = false
    return
  }
  shopTimer = setTimeout(async () => {
    try {
      editForm.shopSuggestions = (await api.searchShops(q)) ?? []
      editForm.showShopSuggestions = true
    } catch {
      editForm.shopSuggestions = []
      editForm.showShopSuggestions = false
    }
  }, 250)
}

function pickShop(shop) {
  editForm.shop_name = shop.name
  editForm.shop_id = shop.id
  editForm.shopSuggestions = []
  editForm.showShopSuggestions = false
}

async function resolveShopId() {
  const name = editForm.shop_name.trim()
  if (!name) throw new Error('Shop is required')
  if (editForm.shop_id) return editForm.shop_id
  try {
    const shop = await api.createShop(name)
    return shop.id
  } catch (e) {
    if (/already exists/i.test(e.message)) {
      const all = await api.getShops()
      const found = (all ?? []).find((s) => s.name.toLowerCase() === name.toLowerCase())
      if (found) return found.id
    }
    throw e
  }
}

async function saveEdit() {
  saving.value = true
  error.value = ''
  try {
    const amount = parseWholeAmount(editForm.amount)
    if (Number.isNaN(amount)) {
      throw new Error('Amount must be a whole number (e.g. 100, 200)')
    }

    const a = parseFloat(editForm.armin_share)
    const r = parseFloat(editForm.ramin_share)
    if (Number.isNaN(a) || Number.isNaN(r) || Math.abs(a + r - 1) > 0.001) {
      throw new Error('Shares must sum to 1')
    }

    const gregorianDate = jalaliToGregorian(editForm.date)
    if (!gregorianDate) throw new Error('Invalid date')

    const name = editForm.name.trim()
    if (!name) throw new Error('Item is required')

    if (!duplicateAcknowledged.value) {
      const result = await api.checkDuplicateExpense({
        name,
        date: gregorianDate,
        exclude_id: editingId.value,
      })
      if (result?.exists) {
        duplicateWarning.value =
          'Same item already recorded on this date. Save again to continue.'
        duplicateAcknowledged.value = true
        return
      }
    }

    const shopId = await resolveShopId()
    await api.updateExpense(editingId.value, {
      person_id: Number(editForm.person_id),
      shop_id: Number(shopId),
      date: gregorianDate,
      name,
      amount,
      shares: [
        { person_id: 1, share: parseFloat(editForm.armin_share) || 0 },
        { person_id: 2, share: parseFloat(editForm.ramin_share) || 0 },
      ],
    })
    editingId.value = null
    clearDuplicateWarning()
    await loadData()
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

async function deleteExpense(id) {
  if (!confirm('Delete this expense?')) return
  try {
    await api.deleteExpense(id)
    await loadData()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(loadData)
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-white">Expenses</h2>
      <button
        class="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white hover:bg-sky-500"
        @click="router.push('/expenses/add')"
      >
        + Add
      </button>
    </div>

    <div
      v-if="error"
      class="rounded-lg border border-red-900 bg-red-950/50 px-4 py-3 text-sm text-red-300"
    >
      {{ error }}
    </div>

    <div
      v-if="duplicateWarning"
      class="rounded-lg border border-amber-800 bg-amber-950/50 px-4 py-3 text-sm text-amber-200"
    >
      {{ duplicateWarning }}
    </div>

    <section class="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <p class="text-sm text-zinc-400">
          {{ expenses.length }} records · {{ formatMoney(grandTotal) }} total
        </p>
        <div class="grid gap-2 sm:grid-cols-3">
          <select
            v-model="filters.person_id"
            class="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
            @change="loadData"
          >
            <option value="">All people</option>
            <option v-for="p in persons" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <JalaliDateInput
            v-model="filters.from_date"
            :input-class="'w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100'"
            @change="loadData"
          />
          <JalaliDateInput
            v-model="filters.to_date"
            :input-class="'w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100'"
            @change="loadData"
          />
        </div>
      </div>

      <div v-if="loading" class="py-12 text-center text-zinc-500">Loading…</div>
      <div v-else-if="expenses.length === 0" class="py-12 text-center text-zinc-500">
        No expenses yet.
      </div>
      <div v-else class="mt-4 overflow-x-auto">
        <table class="min-w-full divide-y divide-zinc-800 text-left text-sm">
          <thead class="text-xs uppercase tracking-wide text-zinc-500">
            <tr>
              <th class="px-3 py-2 font-medium">Person</th>
              <th class="px-3 py-2 font-medium">Item</th>
              <th class="px-3 py-2 font-medium">Shop</th>
              <th class="px-3 py-2 font-medium">Date</th>
              <th class="px-3 py-2 font-medium">Share (Armin)</th>
              <th class="px-3 py-2 font-medium">Share (Ramin)</th>
              <th class="px-3 py-2 font-medium">Amount</th>
              <th class="px-3 py-2 font-medium"></th>
            </tr>
          </thead>
          <tbody class="divide-y divide-zinc-800">
            <tr v-for="exp in expenses" :key="exp.id" class="align-top text-zinc-200">
              <td colspan="8" v-if="editingId === exp.id" class="px-3 py-3">
                <div class="space-y-3 rounded-lg border border-zinc-700 bg-zinc-900 p-4">
                  <div class="grid gap-3 sm:grid-cols-3">
                    <label class="block text-sm">
                      <span class="text-zinc-400">Person</span>
                      <select
                        v-model="editForm.person_id"
                        class="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                      >
                        <option v-for="p in persons" :key="p.id" :value="p.id">{{ p.name }}</option>
                      </select>
                    </label>
                    <JalaliDateInput
                      v-model="editForm.date"
                      label="Date"
                      required
                      @update:model-value="clearDuplicateWarning"
                    />
                    <div class="relative block text-sm">
                      <span class="text-zinc-400">Shop</span>
                      <input
                        v-model="editForm.shop_name"
                        type="text"
                        class="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                        @input="onShopInput"
                        @blur="editForm.showShopSuggestions = false"
                      />
                      <ul
                        v-if="editForm.showShopSuggestions && editForm.shopSuggestions.length"
                        class="absolute z-10 mt-1 max-h-40 w-full overflow-auto rounded border border-zinc-700 bg-zinc-950 shadow-lg"
                      >
                        <li
                          v-for="s in editForm.shopSuggestions"
                          :key="s.id"
                          class="cursor-pointer px-3 py-2 text-sm hover:bg-zinc-800"
                          @mousedown.prevent="pickShop(s)"
                        >
                          {{ s.name }}
                        </li>
                      </ul>
                    </div>
                  </div>
                  <div class="grid gap-2 sm:grid-cols-4">
                    <input
                      v-model="editForm.name"
                      type="text"
                      placeholder="Item"
                      class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                      @input="clearDuplicateWarning"
                    />
                    <input
                      v-model="editForm.amount"
                      type="text"
                      inputmode="numeric"
                      pattern="[0-9]*"
                      placeholder="Amount"
                      class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                    />
                    <input
                      :value="editForm.armin_share"
                      type="number"
                      step="0.1"
                      min="0"
                      max="1"
                      placeholder="Armin share"
                      class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                      @input="setEditShare('armin', $event.target.value)"
                    />
                    <input
                      :value="editForm.ramin_share"
                      type="number"
                      step="0.1"
                      min="0"
                      max="1"
                      placeholder="Ramin share"
                      class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                      @input="setEditShare('ramin', $event.target.value)"
                    />
                  </div>
                  <div class="flex gap-3">
                    <button
                      type="button"
                      class="rounded-lg bg-sky-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-sky-500 disabled:opacity-60"
                      :disabled="saving"
                      @click="saveEdit"
                    >
                      {{
                        saving
                          ? 'Saving…'
                          : duplicateAcknowledged && duplicateWarning
                            ? 'Save anyway'
                            : 'Save'
                      }}
                    </button>
                    <button
                      type="button"
                      class="text-sm text-zinc-400 hover:text-zinc-200"
                      @click="cancelEdit"
                    >
                      Cancel
                    </button>
                  </div>
                </div>
              </td>
              <template v-else>
                <td class="whitespace-nowrap px-3 py-3">{{ exp.person_name }}</td>
                <td class="px-3 py-3">{{ exp.name }}</td>
                <td class="px-3 py-3">{{ exp.shop_name }}</td>
                <td class="whitespace-nowrap px-3 py-3">{{ toJalaliDisplay(exp.date) }}</td>
                <td class="whitespace-nowrap px-3 py-3">{{ shareOf(exp, 1) }}</td>
                <td class="whitespace-nowrap px-3 py-3">{{ shareOf(exp, 2) }}</td>
                <td class="whitespace-nowrap px-3 py-3 font-medium text-white">
                  {{ formatMoney(exp.amount) }}
                </td>
                <td class="whitespace-nowrap px-3 py-3 text-right">
                  <button class="text-sky-400 hover:text-sky-300" @click="startEdit(exp)">
                    Edit
                  </button>
                  <button
                    class="ml-3 text-red-400 hover:text-red-300"
                    @click="deleteExpense(exp.id)"
                  >
                    Delete
                  </button>
                </td>
              </template>
            </tr>
          </tbody>
        </table>
      </div>
    </section>
  </div>
</template>
