<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { jalaliToGregorian, todayJalali } from '../utils/dates'
import { formatMoney } from '../utils/money'
import JalaliDateInput from '../components/JalaliDateInput.vue'

const router = useRouter()
const persons = ref([])
const saving = ref(false)
const error = ref('')
const duplicateWarning = ref('')
const duplicateAcknowledged = ref(false)

const form = reactive({
  person_id: 1,
  shop_name: '',
  shop_id: null,
  shopSuggestions: [],
  showShopSuggestions: false,
  date: todayJalali(),
  items: [emptyLine()],
})

let shopTimer = null
let itemTimers = {}

function emptyLine() {
  return {
    name: '',
    amount: '',
    armin_share: 0.5,
    ramin_share: 0.5,
    suggestions: [],
    showSuggestions: false,
  }
}

const lineTotal = computed(() =>
  form.items.reduce((sum, item) => sum + (parseWholeAmount(item.amount) || 0), 0)
)

const sharesValid = computed(() =>
  form.items.every((item) => {
    const a = parseFloat(item.armin_share)
    const r = parseFloat(item.ramin_share)
    if (Number.isNaN(a) || Number.isNaN(r)) return false
    return Math.abs(a + r - 1) <= 0.001
  })
)

function parseWholeAmount(value) {
  if (value === '' || value === null || value === undefined) return NaN
  const text = String(value).trim()
  if (!/^\d+$/.test(text)) return NaN
  return Number(text)
}

function setShare(item, person, value) {
  const n = parseFloat(value)
  if (Number.isNaN(n)) {
    if (person === 'armin') item.armin_share = value
    else item.ramin_share = value
    return
  }
  const clamped = Math.min(1, Math.max(0, n))
  if (person === 'armin') {
    item.armin_share = clamped
    item.ramin_share = Math.round((1 - clamped) * 1000) / 1000
  } else {
    item.ramin_share = clamped
    item.armin_share = Math.round((1 - clamped) * 1000) / 1000
  }
}

function addLineItem() {
  form.items.push(emptyLine())
  clearDuplicateWarning()
}

function removeLineItem(index) {
  if (form.items.length > 1) form.items.splice(index, 1)
  clearDuplicateWarning()
}

function clearDuplicateWarning() {
  duplicateWarning.value = ''
  duplicateAcknowledged.value = false
}

function onShopInput() {
  form.shop_id = null
  clearTimeout(shopTimer)
  const q = form.shop_name.trim()
  if ([...q].length < 3) {
    form.shopSuggestions = []
    form.showShopSuggestions = false
    return
  }
  shopTimer = setTimeout(async () => {
    try {
      form.shopSuggestions = (await api.searchShops(q)) ?? []
      form.showShopSuggestions = true
    } catch {
      form.shopSuggestions = []
      form.showShopSuggestions = false
    }
  }, 250)
}

function pickShop(shop) {
  form.shop_name = shop.name
  form.shop_id = shop.id
  form.shopSuggestions = []
  form.showShopSuggestions = false
}

function hideShopSuggestions() {
  setTimeout(() => {
    form.showShopSuggestions = false
  }, 150)
}

function onItemNameInput(index) {
  clearDuplicateWarning()
  const item = form.items[index]
  clearTimeout(itemTimers[index])
  const q = item.name.trim()
  if ([...q].length < 3) {
    item.suggestions = []
    item.showSuggestions = false
    return
  }
  itemTimers[index] = setTimeout(async () => {
    try {
      item.suggestions = (await api.searchItems(q)) ?? []
      item.showSuggestions = true
    } catch {
      item.suggestions = []
      item.showSuggestions = false
    }
  }, 250)
}

function pickItemSuggestion(index, name) {
  const item = form.items[index]
  item.name = name
  item.suggestions = []
  item.showSuggestions = false
  clearDuplicateWarning()
}

function hideItemSuggestions(index) {
  setTimeout(() => {
    if (form.items[index]) form.items[index].showSuggestions = false
  }, 150)
}

async function resolveShopId() {
  const name = form.shop_name.trim()
  if (!name) throw new Error('Shop is required')
  if (form.shop_id) return form.shop_id
  try {
    const shop = await api.createShop(name)
    return shop.id
  } catch (e) {
    if (/already exists/i.test(e.message)) {
      const matches = await api.searchShops(name)
      const exact = (matches ?? []).find((s) => s.name.toLowerCase() === name.toLowerCase())
      if (exact) return exact.id
      const all = await api.getShops()
      const found = (all ?? []).find((s) => s.name.toLowerCase() === name.toLowerCase())
      if (found) return found.id
    }
    throw e
  }
}

async function findDuplicateNames(gregorianDate) {
  const duplicates = []
  const seen = new Set()
  for (const item of form.items) {
    const name = item.name.trim()
    if (!name || seen.has(name.toLowerCase())) continue
    seen.add(name.toLowerCase())
    const result = await api.checkDuplicateExpense({ name, date: gregorianDate })
    if (result?.exists) duplicates.push(name)
  }
  return duplicates
}

async function submitExpense() {
  saving.value = true
  error.value = ''
  try {
    if (!sharesValid.value) throw new Error('Each item shares must sum to 1')

    for (const item of form.items) {
      const amount = parseWholeAmount(item.amount)
      if (Number.isNaN(amount)) {
        throw new Error('Amount must be a whole number (e.g. 100, 200)')
      }
    }

    const gregorianDate = jalaliToGregorian(form.date)
    if (!gregorianDate) throw new Error('Invalid date')

    if (!duplicateAcknowledged.value) {
      const duplicates = await findDuplicateNames(gregorianDate)
      if (duplicates.length) {
        duplicateWarning.value = `Same item already recorded on this date: ${duplicates.join(', ')}. Save again to continue.`
        duplicateAcknowledged.value = true
        return
      }
    }

    const shopId = await resolveShopId()

    await api.createExpenses({
      person_id: Number(form.person_id),
      shop_id: Number(shopId),
      date: gregorianDate,
      items: form.items.map((item) => ({
        name: item.name.trim(),
        amount: parseWholeAmount(item.amount),
        shares: [
          { person_id: 1, share: parseFloat(item.armin_share) || 0 },
          { person_id: 2, share: parseFloat(item.ramin_share) || 0 },
        ],
      })),
    })

    router.push('/expenses/list')
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    persons.value = (await api.getPersons()) ?? []
    if (persons.value.length) form.person_id = persons.value[0].id
  } catch (e) {
    error.value = e.message
  }
})
</script>

<template>
  <div class="space-y-4">
    <h2 class="text-lg font-semibold text-white">Add expense</h2>

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

    <form class="space-y-4 rounded-xl border border-zinc-800 bg-zinc-950 p-5" @submit.prevent="submitExpense">
      <div class="grid gap-4 sm:grid-cols-3">
        <label class="block text-sm">
          <span class="font-medium text-zinc-300">Paid by</span>
          <select
            v-model="form.person_id"
            class="mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100"
            required
          >
            <option v-for="p in persons" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
        </label>

        <JalaliDateInput v-model="form.date" label="Date" required @update:model-value="clearDuplicateWarning" />

        <div class="relative block text-sm">
          <span class="font-medium text-zinc-300">Shop</span>
          <input
            v-model="form.shop_name"
            type="text"
            placeholder="Search or create shop"
            class="mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 placeholder:text-zinc-500"
            required
            autocomplete="off"
            @input="onShopInput"
            @focus="onShopInput"
            @blur="hideShopSuggestions"
          />
          <ul
            v-if="form.showShopSuggestions && form.shopSuggestions.length"
            class="absolute z-10 mt-1 max-h-40 w-full overflow-auto rounded border border-zinc-700 bg-zinc-950 shadow-lg"
          >
            <li
              v-for="s in form.shopSuggestions"
              :key="s.id"
              class="cursor-pointer px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800"
              @mousedown.prevent="pickShop(s)"
            >
              {{ s.name }}
            </li>
          </ul>
        </div>
      </div>

      <div>
        <div class="mb-2 flex items-center justify-between">
          <span class="text-sm font-medium text-zinc-300">Items</span>
          <button
            type="button"
            class="text-sm font-medium text-sky-400 hover:text-sky-300"
            @click="addLineItem"
          >
            + Add item
          </button>
        </div>
        <div class="space-y-2">
          <div
            v-for="(item, index) in form.items"
            :key="index"
            class="grid gap-2 rounded-lg bg-zinc-900 p-3 sm:grid-cols-12"
          >
            <div class="relative sm:col-span-4">
              <input
                v-model="item.name"
                type="text"
                placeholder="Item"
                class="w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100 placeholder:text-zinc-500"
                required
                autocomplete="off"
                @input="onItemNameInput(index)"
                @focus="onItemNameInput(index)"
                @blur="hideItemSuggestions(index)"
              />
              <ul
                v-if="item.showSuggestions && item.suggestions.length"
                class="absolute z-10 mt-1 max-h-40 w-full overflow-auto rounded border border-zinc-700 bg-zinc-950 shadow-lg"
              >
                <li
                  v-for="s in item.suggestions"
                  :key="s.id"
                  class="cursor-pointer px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800"
                  @mousedown.prevent="pickItemSuggestion(index, s.name)"
                >
                  {{ s.name }}
                </li>
              </ul>
            </div>
            <input
              v-model="item.amount"
              type="text"
              inputmode="numeric"
              pattern="[0-9]*"
              placeholder="Amount"
              class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100 sm:col-span-2"
              required
            />
            <input
              :value="item.armin_share"
              type="number"
              step="0.1"
              min="0"
              max="1"
              placeholder="Armin share"
              class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100 sm:col-span-2"
              required
              @input="setShare(item, 'armin', $event.target.value)"
            />
            <input
              :value="item.ramin_share"
              type="number"
              step="0.1"
              min="0"
              max="1"
              placeholder="Ramin share"
              class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100 sm:col-span-2"
              required
              @input="setShare(item, 'ramin', $event.target.value)"
            />
            <button
              type="button"
              class="text-sm text-red-400 hover:text-red-300 sm:col-span-2"
              :disabled="form.items.length === 1"
              @click="removeLineItem(index)"
            >
              Remove
            </button>
          </div>
        </div>
        <p class="mt-2 text-right text-sm text-zinc-400">
          Total: <span class="font-semibold text-zinc-100">{{ formatMoney(lineTotal) }}</span>
        </p>
      </div>

      <button
        type="submit"
        :disabled="saving || !sharesValid"
        class="w-full rounded-lg bg-sky-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-sky-500 disabled:opacity-60 sm:w-auto"
      >
        {{ saving ? 'Saving…' : duplicateAcknowledged && duplicateWarning ? 'Save anyway' : 'Save' }}
      </button>
    </form>
  </div>
</template>
