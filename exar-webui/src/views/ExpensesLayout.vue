<template>
  <div class="space-y-4">
    <div class="flex flex-wrap items-center gap-2 border-b border-zinc-800 pb-3">
      <router-link
        v-for="tab in tabs"
        :key="tab.to"
        :to="tab.to"
        class="rounded-lg px-3 py-1.5 text-sm font-medium text-zinc-400 hover:bg-zinc-900 hover:text-zinc-100"
        active-class="!bg-zinc-900 !text-sky-400"
      >
        {{ tab.label }}
      </router-link>
    </div>
    <router-view />
  </div>
</template>

<script setup>
const tabs = [
  { to: '/expenses/list', label: 'List' },
  { to: '/expenses/add', label: 'Add' },
]
</script>
