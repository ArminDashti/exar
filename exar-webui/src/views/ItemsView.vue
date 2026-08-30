<script setup>
import { onMounted, ref } from 'vue'
import { api } from '../api'

const items = ref([])
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
    items.value = (await api.getItems()) ?? []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function createItem() {
  const name = newName.value.trim()
  if (!name) return
  saving.value = true
  error.value = ''
  try {
    await api.createItem(name)
    newName.value = ''
    await load()
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

function startEdit(item) {
  editingId.value = item.id
  editingName.value = item.name
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
    await api.updateItem(id, name)
    cancelEdit()
    await load()
  } catch (e) {
    error.value = e.message
  } finally {
    editSaving.value = false
  }
}

async function removeItem(id) {
  if (!confirm('Delete this item from the catalog?')) return
  error.value = ''
  try {
    await api.deleteItem(id)
    await load()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(load)
</script>

<template>
  <div class="space-y-6">
    <h2 class="text-lg font-semibold text-white">Items</h2>
    <p class="text-sm text-zinc-400">
      Catalog of item names used when adding expenses. Renaming an item updates it on all
      expenses. Search starts after 3 characters.
    </p>

    <div
      v-if="error"
      class="rounded-lg border border-red-900 bg-red-950/50 px-4 py-3 text-sm text-red-300"
    >
      {{ error }}
    </div>

    <form
      class="flex flex-col gap-3 rounded-xl border border-zinc-800 bg-zinc-950 p-5 sm:flex-row"
      @submit.prevent="createItem"
    >
      <input
        v-model="newName"
        type="text"
        placeholder="New item name"
        class="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500"
        required
      />
      <button
        type="submit"
        :disabled="saving"
        class="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white hover:bg-sky-500 disabled:opacity-60"
      >
        {{ saving ? 'Adding…' : 'Add item' }}
      </button>
    </form>

    <section class="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
      <div v-if="loading" class="py-8 text-center text-zinc-500">Loading…</div>
      <div v-else-if="items.length === 0" class="py-8 text-center text-zinc-500">
        No items yet.
      </div>
      <ul v-else class="divide-y divide-zinc-800">
        <li
          v-for="item in items"
          :key="item.id"
          class="flex flex-col gap-2 py-3 sm:flex-row sm:items-center sm:justify-between"
        >
          <template v-if="editingId === item.id">
            <input
              v-model="editingName"
              type="text"
              class="flex-1 rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-1.5 text-sm text-zinc-100"
              @keydown.enter.prevent="saveEdit(item.id)"
              @keydown.escape="cancelEdit"
            />
            <div class="flex gap-3">
              <button
                class="text-sm text-sky-400 hover:text-sky-300"
                :disabled="editSaving"
                @click="saveEdit(item.id)"
              >
                {{ editSaving ? 'Saving…' : 'Save' }}
              </button>
              <button class="text-sm text-zinc-400 hover:text-zinc-300" @click="cancelEdit">
                Cancel
              </button>
            </div>
          </template>
          <template v-else>
            <span class="text-sm text-zinc-100">{{ item.name }}</span>
            <div class="flex gap-3">
              <button
                class="text-sm text-sky-400 hover:text-sky-300"
                @click="startEdit(item)"
              >
                Edit
              </button>
              <button
                class="text-sm text-red-400 hover:text-red-300"
                @click="removeItem(item.id)"
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
