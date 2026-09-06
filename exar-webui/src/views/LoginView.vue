<template>
  <div class="mx-auto flex min-h-[70vh] max-w-md flex-col justify-center">
    <div class="rounded-2xl border border-zinc-800 bg-zinc-950 p-6 shadow-xl">
      <h2 class="text-xl font-semibold text-white">Sign in</h2>
      <p class="mt-1 text-sm text-zinc-400">Use your exar username and password</p>

      <form class="mt-6 space-y-4" @submit.prevent="submit">
        <label class="block text-sm text-zinc-300">
          Username
          <input
            v-model.trim="username"
            autocomplete="username"
            class="mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 outline-none focus:border-sky-500"
            required
          />
        </label>

        <label class="block text-sm text-zinc-300">
          Password
          <input
            v-model="password"
            type="password"
            autocomplete="current-password"
            class="mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100 outline-none focus:border-sky-500"
            required
          />
        </label>

        <p v-if="error" class="text-sm text-red-400">{{ error }}</p>

        <button
          type="submit"
          class="w-full rounded-lg bg-sky-600 px-4 py-2 text-sm font-medium text-white hover:bg-sky-500 disabled:opacity-60"
          :disabled="loading"
        >
          {{ loading ? 'Signing in…' : 'Sign in' }}
        </button>
      </form>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { api } from '../api'
import { setSession } from '../auth'

const router = useRouter()
const route = useRoute()

const username = ref('')
const password = ref('')
const error = ref('')
const loading = ref(false)

async function submit() {
  error.value = ''
  loading.value = true
  try {
    const data = await api.login(username.value, password.value)
    setSession(data.token, {
      username: data.username,
      person_id: data.person_id,
    })
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/expenses/list'
    await router.replace(redirect || '/expenses/list')
  } catch (err) {
    error.value = err.message || 'Login failed'
  } finally {
    loading.value = false
  }
}
</script>
