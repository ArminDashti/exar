<script setup>
import { onMounted, ref } from 'vue'
import { api } from '../api'

const shops = ref([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const newName = ref('')
const editingId = ref(null)
const editingName = ref('')
const editSaving = ref(false)

async function load() {
  loading.value = true
  error.value = ''
  try {
    shops.value = (await api.getShops()) ?? []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function createShop() {
  const name = newName.value.trim()
  if (!name) return
  saving.value = true
  error.value = ''
  try {
    await api.createShop(name)
    newName.value = ''
    await load()
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

function startEdit(shop) {
  editingId.value = shop.id
  editingName.value = shop.name
}

function cancelEdit() {
  editingId.value = null
  editingName.value = ''
}

async function saveEdit(id) {
  const name = editingName.value.trim()
  if (!name) return
  editSaving.value = true
  error.value = ''
  try {
    await api.updateShop(id, name)
    cancelEdit()
    await load()
  } catch (e) {
    error.value = e.message
  } finally {
    editSaving.value = false
  }
}

async function removeShop(id) {
  if (!confirm('Delete this shop?')) return
  error.value = ''
  try {
    await api.deleteShop(id)
    await load()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <h2 class="text-lg font-semibold text-white">Shops</h2>

    <div
      v-if="error"
      class="rounded-lg border border-red-900 bg-red-950/50 px-4 py-3 text-sm text-red-300"
    >
      {{ error }}
    </div>

    <form
      class="flex flex-col gap-3 rounded-xl border border-zinc-800 bg-zinc-950 p-5 sm:flex-row"
      @submit.prevent="createShop"
    >
      <input
        v-model="newName"
        type="text"
        placeholder="New shop name"
        class="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500"
        required
      />
      <button
        type="submit"
        :disabled="saving"
        class="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white hover:bg-sky-500 disabled:opacity-60"
      >
        {{ saving ? 'Adding…' : 'Add shop' }}
      </button>
    </form>

    <section class="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
      <div v-if="loading" class="py-8 text-center text-zinc-500">Loading…</div>
      <div v-else-if="shops.length === 0" class="py-8 text-center text-zinc-500">
        No shops yet.
      </div>
      <ul v-else class="divide-y divide-zinc-800">
        <li
          v-for="shop in shops"
          :key="shop.id"
          class="flex flex-col gap-2 py-3 sm:flex-row sm:items-center sm:justify-between"
        >
          <template v-if="editingId === shop.id">
            <input
              v-model="editingName"
              type="text"
              class="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100"
              @keydown.enter.prevent="saveEdit(shop.id)"
              @keydown.escape="cancelEdit"
            />
            <div class="flex gap-3">
              <button
                class="text-sm text-sky-400 hover:text-sky-300"
                :disabled="editSaving"
                @click="saveEdit(shop.id)"
              >
                {{ editSaving ? 'Saving…' : 'Save' }}
              </button>
              <button class="text-sm text-zinc-400 hover:text-zinc-300" @click="cancelEdit">
                Cancel
              </button>
            </div>
          </template>
          <template v-else>
            <span class="text-sm text-zinc-100">{{ shop.name }}</span>
            <div class="flex gap-3">
              <button
                class="text-sm text-sky-400 hover:text-sky-300"
                @click="startEdit(shop)"
              >
                Edit
              </button>
              <button
                class="text-sm text-red-400 hover:text-red-300"
                @click="removeShop(shop.id)"
              >
                Delete
              </button>
            </div>
          </template>
        </li>
      </ul>
    </section>
  </div>
</template>
