<script setup>
import { onMounted, reactive, ref } from 'vue'
import { api } from '../api'
import { jalaliToGregorian } from '../utils/dates'
import { formatMoney } from '../utils/money'
import JalaliDateInput from '../components/JalaliDateInput.vue'

const loading = ref(true)
const error = ref('')
const stats = ref({ total: 0, by_person: [], by_shop: [] })

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
    if (filters.from_date && !from) throw new Error('Invalid from date (use YYYY/MM/DD Jalali)')
    if (filters.to_date && !to) throw new Error('Invalid to date (use YYYY/MM/DD Jalali)')

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
        :input-class="'mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500'"
      />
      <JalaliDateInput
        v-model="filters.to_date"
        label="To"
        :input-class="'mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500'"
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
    <template v-else>
      <section class="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
        <h3 class="text-sm font-medium text-zinc-400">Total spend</h3>
        <p class="mt-2 text-3xl font-semibold text-white">{{ formatMoney(stats.total) }}</p>
      </section>

      <section class="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
        <h3 class="mb-3 text-sm font-medium text-zinc-400">By person</h3>
        <ul class="divide-y divide-zinc-800">
          <li
            v-for="row in stats.by_person"
            :key="row.person_id"
            class="flex items-center justify-between py-3"
          >
            <span class="text-sm text-zinc-100">{{ row.person_name }}</span>
            <span class="text-sm font-semibold text-white">{{ formatMoney(row.total) }}</span>
          </li>
        </ul>
      </section>

      <section class="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
        <h3 class="mb-3 text-sm font-medium text-zinc-400">By shop</h3>
        <div v-if="!stats.by_shop?.length" class="py-6 text-center text-sm text-zinc-500">
          No shop spend in this range.
        </div>
        <ul v-else class="divide-y divide-zinc-800">
          <li
            v-for="row in stats.by_shop"
            :key="row.shop_id"
            class="flex items-center justify-between py-3"
          >
            <span class="text-sm text-zinc-100">{{ row.shop_name }}</span>
            <span class="text-sm font-semibold text-white">{{ formatMoney(row.total) }}</span>
          </li>
        </ul>
      </section>
    </template>
  </div>
</template>
