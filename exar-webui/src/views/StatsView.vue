<script setup>
import { onMounted, reactive, ref } from 'vue'
import { api } from '../api'
import { jalaliToGregorian } from '../utils/dates'
import { formatMoney } from '../utils/money'
import JalaliDateInput from '../components/JalaliDateInput.vue'

const loading = ref(true)
const error = ref('')
const stats = ref({ by_month: [] })

const filters = reactive({
  from_date: '',
  to_date: '',
})

async function load() {
  loading.value = true
  error.value = ''
  try {
    const from = filters.from_date ? jalaliToGregorian(filters.from_date) : undefined
    const to = filters.to_date ? jalaliToGregorian(filters.to_date) : undefined
    if (filters.from_date && !from) throw new Error('Invalid from date')
    if (filters.to_date && !to) throw new Error('Invalid to date')

    stats.value = await api.getStats({
      from_date: from,
      to_date: to,
    })
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <h2 class="text-lg font-semibold text-white">Stats</h2>

    <div class="grid gap-2 rounded-xl border border-zinc-800 bg-zinc-950 p-5 sm:grid-cols-3">
      <JalaliDateInput
        v-model="filters.from_date"
        label="From"
        :input-class="'mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100'"
      />
      <JalaliDateInput
        v-model="filters.to_date"
        label="To"
        :input-class="'mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100'"
      />
      <div class="flex items-end">
        <button
          class="w-full rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white hover:bg-sky-500"
          @click="load"
        >
          Apply
        </button>
      </div>
    </div>

    <div
      v-if="error"
      class="rounded-lg border border-red-900 bg-red-950/50 px-4 py-3 text-sm text-red-300"
    >
      {{ error }}
    </div>

    <div v-if="loading" class="py-12 text-center text-zinc-500">Loading…</div>
    <section v-else class="overflow-x-auto rounded-xl border border-zinc-800 bg-zinc-950">
      <table class="min-w-full text-left text-sm">
        <thead class="border-b border-zinc-800 text-xs uppercase text-zinc-500">
          <tr>
            <th class="px-4 py-3 font-medium">Month</th>
            <th class="px-4 py-3 font-medium">Armin</th>
            <th class="px-4 py-3 font-medium">Ramin</th>
            <th class="px-4 py-3 font-medium">Total</th>
            <th class="px-4 py-3 font-medium">Armin share</th>
            <th class="px-4 py-3 font-medium">Ramin share</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-zinc-800">
          <tr v-if="!stats.by_month?.length">
            <td colspan="6" class="px-4 py-10 text-center text-zinc-500">No data in this range.</td>
          </tr>
          <tr v-for="row in stats.by_month" :key="row.month" class="text-zinc-200">
            <td class="px-4 py-3 font-medium text-white">{{ row.month }}</td>
            <td class="px-4 py-3">{{ formatMoney(row.armin) }}</td>
            <td class="px-4 py-3">{{ formatMoney(row.ramin) }}</td>
            <td class="px-4 py-3">{{ formatMoney(row.total) }}</td>
            <td class="px-4 py-3">{{ formatMoney(row.armin_share) }}</td>
            <td class="px-4 py-3">{{ formatMoney(row.ramin_share) }}</td>
          </tr>
        </tbody>
      </table>
    </section>
  </div>
</template>
