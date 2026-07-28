<template>
  <div class="flex min-h-screen flex-col bg-black text-zinc-100">
    <header class="border-b border-zinc-800 bg-zinc-950">
      <div class="mx-auto flex max-w-5xl flex-col gap-4 px-4 py-5 sm:flex-row sm:items-center sm:justify-between sm:px-6">
        <div>
          <h1 class="text-2xl font-bold tracking-tight text-white">exar</h1>
          <p class="mt-1 text-sm text-zinc-400">Track who spent what, where, and when</p>
        </div>
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
      </div>
    </header>

    <main class="mx-auto w-full max-w-5xl flex-1 px-4 py-6 sm:px-6">
      <router-view />
    </main>

    <footer class="border-t border-zinc-800 bg-zinc-950 py-4 text-center text-xs text-zinc-500">
      © {{ year }} Dashti Technologies (Armin Dashti)
    </footer>
  </div>
</template>

<script setup>
const year = new Date().getFullYear()
const links = [
  { to: '/expenses', label: 'Expenses' },
  { to: '/shops', label: 'Shops' },
  { to: '/items', label: 'Items' },
  { to: '/stats', label: 'Stats' },
]
</script>
