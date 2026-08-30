<template>
  <div ref="root" class="relative block text-sm">
    <span v-if="label" class="font-medium text-zinc-300">{{ label }}</span>
    <button
      type="button"
      :class="[inputClass, label ? 'mt-1' : '']"
      class="flex w-full items-center justify-between text-left"
      @click.stop="toggle"
    >
      <span :class="modelValue ? 'text-zinc-100' : 'text-zinc-500'">
        {{ modelValue || 'YYYY/MM/DD' }}
      </span>
      <svg
        class="h-4 w-4 shrink-0 text-zinc-400"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="2"
      >
        <rect x="3" y="4" width="18" height="18" rx="2" />
        <path d="M16 2v4M8 2v4M3 10h18" />
      </svg>
    </button>
    <!-- Hidden input so required works in native forms -->
    <input
      v-if="required"
      type="text"
      class="sr-only"
      tabindex="-1"
      :value="modelValue"
      required
      @focus="open = true"
    />

    <div
      v-if="open"
      class="absolute z-20 mt-1 w-72 rounded-lg border border-zinc-700 bg-zinc-950 p-3 shadow-xl"
      @click.stop
    >
      <div class="mb-2 flex items-center justify-between">
        <button
          type="button"
          class="rounded px-2 py-1 text-zinc-300 hover:bg-zinc-800"
          @click="prevMonth"
        >
          ‹
        </button>
        <span class="text-sm font-medium text-zinc-100">
          {{ viewYear }}/{{ pad(viewMonth) }}
        </span>
        <button
          type="button"
          class="rounded px-2 py-1 text-zinc-300 hover:bg-zinc-800"
          @click="nextMonth"
        >
          ›
        </button>
      </div>
      <div class="mb-1 grid grid-cols-7 gap-1 text-center text-[10px] text-zinc-500">
        <span v-for="d in weekdays" :key="d">{{ d }}</span>
      </div>
      <div class="grid grid-cols-7 gap-1">
        <button
          v-for="(cell, idx) in cells"
          :key="idx"
          type="button"
          class="rounded py-1.5 text-xs"
          :class="cellClasses(cell)"
          :disabled="!cell.day"
          @click="selectDay(cell.day)"
        >
          {{ cell.day || '' }}
        </button>
      </div>
      <div class="mt-2 flex justify-between">
        <button type="button" class="text-xs text-sky-400 hover:text-sky-300" @click="pickToday">
          Today
        </button>
        <button type="button" class="text-xs text-zinc-500 hover:text-zinc-300" @click="open = false">
          Close
        </button>
      </div>
    </div>
  </div>
</template>

<script setup>
import { computed, onMounted, onUnmounted, ref, watch } from 'vue'
import { isValidJalaaliDate, jalaaliMonthLength, toGregorian } from 'jalaali-js'
import { todayJalali } from '../utils/dates'

const props = defineProps({
  modelValue: { type: String, default: '' },
  label: { type: String, default: '' },
  required: { type: Boolean, default: false },
  inputClass: {
    type: String,
    default:
      'w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100',
  },
})

const emit = defineEmits(['update:modelValue', 'change'])

const root = ref(null)
const open = ref(false)
const viewYear = ref(1400)
const viewMonth = ref(1)
const weekdays = ['Sh', 'Ye', 'Do', 'Se', 'Ch', 'Pa', 'Jo']

function pad(n) {
  return String(n).padStart(2, '0')
}

function parseJalali(value) {
  if (!value) return null
  const parts = String(value).trim().replace(/-/g, '/').split('/').map(Number)
  if (parts.length !== 3) return null
  const [jy, jm, jd] = parts
  if (!isValidJalaaliDate(jy, jm, jd)) return null
  return { jy, jm, jd }
}

function syncViewFromValue() {
  const parsed = parseJalali(props.modelValue) || parseJalali(todayJalali())
  if (parsed) {
    viewYear.value = parsed.jy
    viewMonth.value = parsed.jm
  }
}

watch(
  () => props.modelValue,
  () => {
    if (!open.value) syncViewFromValue()
  }
)

const cells = computed(() => {
  const length = jalaaliMonthLength(viewYear.value, viewMonth.value)
  const g = toGregorian(viewYear.value, viewMonth.value, 1)
  const jsDay = new Date(g.gy, g.gm - 1, g.gd).getDay()
  const startOffset = (jsDay + 1) % 7

  const result = []
  for (let i = 0; i < startOffset; i++) result.push({ day: 0 })
  for (let d = 1; d <= length; d++) result.push({ day: d })
  while (result.length % 7 !== 0) result.push({ day: 0 })
  return result
})

function cellClasses(cell) {
  if (!cell.day) return 'cursor-default text-transparent'
  const selected = parseJalali(props.modelValue)
  const isSelected =
    selected &&
    selected.jy === viewYear.value &&
    selected.jm === viewMonth.value &&
    selected.jd === cell.day
  if (isSelected) return 'bg-sky-600 font-semibold text-white'
  return 'text-zinc-200 hover:bg-zinc-800'
}

function toggle() {
  open.value = !open.value
  if (open.value) syncViewFromValue()
}

function prevMonth() {
  if (viewMonth.value === 1) {
    viewMonth.value = 12
    viewYear.value -= 1
  } else {
    viewMonth.value -= 1
  }
}

function nextMonth() {
  if (viewMonth.value === 12) {
    viewMonth.value = 1
    viewYear.value += 1
  } else {
    viewMonth.value += 1
  }
}

function selectDay(day) {
  if (!day) return
  const value = `${viewYear.value}/${pad(viewMonth.value)}/${pad(day)}`
  emit('update:modelValue', value)
  emit('change')
  open.value = false
}

function pickToday() {
  emit('update:modelValue', todayJalali())
  emit('change')
  open.value = false
  syncViewFromValue()
}

function outsideClose(e) {
  if (!open.value) return
  if (root.value && !root.value.contains(e.target)) {
    open.value = false
  }
}

onMounted(() => {
  syncViewFromValue()
  document.addEventListener('click', outsideClose)
})

onUnmounted(() => {
  document.removeEventListener('click', outsideClose)
})
</script>
