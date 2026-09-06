<template>
  <div class="flex min-h-screen flex-col bg-black text-zinc-100">
    <header v-if="showChrome" class="border-b border-zinc-800 bg-zinc-950">
      <div class="mx-auto flex max-w-5xl flex-col gap-4 px-4 py-5 sm:flex-row sm:items-center sm:justify-between sm:px-6">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">exar</h1>
          <p class="mt-1 text-sm text-zinc-400">Track who spent what, where, and when</p>
        </div>
        <div class="flex flex-wrap items-center gap-2">
          <nav class="flex flex-wrap gap-1 text-sm font-medium text-zinc-400">
            <router-link
              v-for="link in links"
              :key="link.to"
              :to="link.to"
              class="rounded-lg px-3 py-2 hover:bg-zinc-900 hover:text-zinc-100"
              active-class="!bg-zinc-900 !text-sky-400"
            >
              {{ link.label }}
            </router-link>
          </nav>
          <button
            type="button"
            class="rounded-lg px-3 py-2 text-sm text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100"
            @click="onLogout"
          >
            Sign out{{ userLabel }}
          </button>
        </div>
      </div>
    </header>

    <main class="mx-auto w-full max-w-5xl flex-1 px-4 py-6 sm:px-6">
      <router-view />
    </main>

    <footer v-if="showChrome" class="border-t border-zinc-800 bg-zinc-950 py-4">
      <div class="mx-auto flex max-w-5xl items-center justify-between px-4 sm:px-6">
        <a
          href="https://github.com/ArminDashti"
          target="_blank"
          rel="noopener noreferrer"
          class="inline-flex items-center gap-2 text-xs text-zinc-400 hover:text-zinc-100"
          aria-label="Armin Dashti on GitHub"
        >
          <svg class="h-4 w-4" viewBox="0 0 24 24" fill="currentColor" aria-hidden="true">
            <path
              fill-rule="evenodd"
              clip-rule="evenodd"
              d="M12 2C6.477 2 2 6.484 2 12.017c0 4.425 2.865 8.18 6.839 9.504.5.092.682-.217.682-.483 0-.237-.008-.868-.013-1.703-2.782.605-3.369-1.343-3.369-1.343-.454-1.158-1.11-1.466-1.11-1.466-.908-.62.069-.608.069-.608 1.003.07 1.531 1.032 1.531 1.032.892 1.53 2.341 1.088 2.91.832.092-.647.35-1.088.636-1.338-2.22-.253-4.555-1.113-4.555-4.951 0-1.093.39-1.988 1.029-2.688-.103-.253-.446-1.272.098-2.65 0 0 .84-.27 2.75 1.026A9.564 9.564 0 0112 6.844c.85.004 1.705.115 2.504.337 1.909-1.296 2.747-1.027 2.747-1.027.546 1.379.202 2.398.1 2.651.64.7 1.028 1.595 1.028 2.688 0 3.848-2.339 4.695-4.566 4.943.359.309.678.92.678 1.855 0 1.338-.012 2.419-.012 2.747 0 .268.18.58.688.482A10.019 10.019 0 0022 12.017C22 6.484 17.522 2 12 2z"
            />
          </svg>
          GitHub
        </a>
        <span class="text-xs text-zinc-500">© {{ year }} Dashti Technologies (Armin Dashti)</span>
      </div>
    </footer>
  </div>
</template>

<script setup>
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import { getUser } from './auth'
import { logoutAndRedirect } from './router'

const route = useRoute()
const year = new Date().getFullYear()
const links = [
  { to: '/expenses', label: 'Expenses' },
  { to: '/shops', label: 'Shops' },
  { to: '/items', label: 'Items' },
  { to: '/stats', label: 'Stats' },
]

const showChrome = computed(() => route.name !== 'login')
const userLabel = computed(() => {
  const user = getUser()
  return user?.username ? ` (${user.username})` : ''
})

function onLogout() {
  logoutAndRedirect()
}
</script>
