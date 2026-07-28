<script setup>
import { computed, onMounted, reactive, ref, watch } from 'vue'
import { api } from '../api'
import { jalaliToGregorian, todayJalali, toJalaliDisplay } from '../utils/dates'
import { formatMoney } from '../utils/money'
import JalaliDateInput from '../components/JalaliDateInput.vue'

const persons = ref([])
const shops = ref([])
const invoices = ref([])
const loading = ref(true)
const saving = ref(false)
const error = ref('')
const showForm = ref(false)
const editingId = ref(null)

const filters = reactive({
  person_id: '',
  from_date: '',
  to_date: '',
})

const form = reactive({
  person_id: 1,
  shop_id: '',
  newShopName: '',
  useNewShop: false,
  date: todayJalali(),
  items: [{ description: '', amount: '', suggestions: [], showSuggestions: false }],
  shares: {},
})

let searchTimers = {}

const grandTotal = computed(() =>
  (invoices.value ?? []).reduce((sum, inv) => sum + inv.total, 0)
)

const lineTotal = computed(() =>
  form.items.reduce((sum, item) => sum + (parseFloat(item.amount) || 0), 0)
)

const shareSum = computed(() =>
  persons.value.reduce((sum, p) => sum + (parseFloat(form.shares[p.id]) || 0), 0)
)

const sharesValid = computed(() => Math.abs(shareSum.value - 1) <= 0.001)

watch(
  persons,
  (list) => {
    for (const p of list) {
      if (form.shares[p.id] === undefined) {
        form.shares[p.id] = list.length ? 1 / list.length : 0.5
      }
    }
  },
  { immediate: true }
)

function resetForm() {
  editingId.value = null
  form.person_id = 1
  form.shop_id = ''
  form.newShopName = ''
  form.useNewShop = false
  form.date = todayJalali()
  form.items = [{ description: '', amount: '', suggestions: [], showSuggestions: false }]
  for (const p of persons.value) {
    form.shares[p.id] = persons.value.length ? 1 / persons.value.length : 0.5
  }
}

function toggleForm() {
  if (showForm.value) {
    showForm.value = false
    resetForm()
  } else {
    resetForm()
    showForm.value = true
  }
}

function gregorianFilter() {
  const from = filters.from_date ? jalaliToGregorian(filters.from_date) : undefined
  const to = filters.to_date ? jalaliToGregorian(filters.to_date) : undefined
  if (filters.from_date && !from) throw new Error('Invalid from date (use YYYY/MM/DD Jalali)')
  if (filters.to_date && !to) throw new Error('Invalid to date (use YYYY/MM/DD Jalali)')
  return { from_date: from || undefined, to_date: to || undefined }
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const dates = gregorianFilter()
    const [p, s, inv] = await Promise.all([
      api.getPersons(),
      api.getShops(),
      api.getInvoices({
        person_id: filters.person_id || undefined,
        ...dates,
      }),
    ])
    persons.value = p ?? []
    shops.value = s ?? []
    invoices.value = inv ?? []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function addLineItem() {
  form.items.push({ description: '', amount: '', suggestions: [], showSuggestions: false })
}

function removeLineItem(index) {
  if (form.items.length > 1) form.items.splice(index, 1)
}

function onItemNameInput(index) {
  const item = form.items[index]
  clearTimeout(searchTimers[index])
  const q = item.description.trim()
  if ([...q].length < 3) {
    item.suggestions = []
    item.showSuggestions = false
    return
  }
  searchTimers[index] = setTimeout(async () => {
    try {
      const results = await api.searchItems(q)
      item.suggestions = results ?? []
      item.showSuggestions = true
    } catch {
      item.suggestions = []
      item.showSuggestions = false
    }
  }, 250)
}

function pickSuggestion(index, name) {
  const item = form.items[index]
  item.description = name
  item.suggestions = []
  item.showSuggestions = false
}

function hideSuggestions(index) {
  setTimeout(() => {
    if (form.items[index]) form.items[index].showSuggestions = false
  }, 150)
}

async function ensureCatalogItems(names) {
  for (const name of names) {
    try {
      await api.createItem(name)
    } catch (e) {
      if (!/already exists/i.test(e.message)) throw e
    }
  }
}

function startEdit(inv) {
  editingId.value = inv.id
  form.person_id = inv.person_id
  form.shop_id = inv.shop_id
  form.newShopName = ''
  form.useNewShop = false
  form.date = toJalaliDisplay(inv.date)
  form.items = (inv.items?.length ? inv.items : [{ description: '', amount: '' }]).map(
    (item) => ({
      description: item.description,
      amount: item.amount,
      suggestions: [],
      showSuggestions: false,
    })
  )
  for (const p of persons.value) {
    const found = inv.shares?.find((s) => s.person_id === p.id)
    form.shares[p.id] = found ? found.share : 0
  }
  showForm.value = true
  window.scrollTo({ top: 0, behavior: 'smooth' })
}

async function submitExpense() {
  saving.value = true
  error.value = ''
  try {
    if (!sharesValid.value) throw new Error('Shares must sum to 1')

    const gregorianDate = jalaliToGregorian(form.date)
    if (!gregorianDate) throw new Error('Invalid date (use YYYY/MM/DD Jalali)')

    let shopId = form.shop_id
    if (form.useNewShop) {
      const shop = await api.createShop(form.newShopName.trim())
      shopId = shop.id
      shops.value.push(shop)
    }
    if (!shopId) throw new Error('Select or create a shop')

    const names = form.items.map((item) => item.description.trim()).filter(Boolean)
    await ensureCatalogItems(names)

    const payload = {
      person_id: Number(form.person_id),
      shop_id: Number(shopId),
      date: gregorianDate,
      items: form.items.map((item) => ({
        description: item.description.trim(),
        amount: parseFloat(item.amount),
      })),
      shares: persons.value.map((p) => ({
        person_id: p.id,
        share: parseFloat(form.shares[p.id]) || 0,
      })),
    }

    if (editingId.value) {
      await api.updateInvoice(editingId.value, payload)
    } else {
      await api.createInvoice(payload)
    }

    showForm.value = false
    resetForm()
    await loadData()
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

async function deleteInvoice(id) {
  if (!confirm('Delete this expense?')) return
  try {
    await api.deleteInvoice(id)
    await loadData()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(loadData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-white">Expenses</h2>
      <button
        class="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white hover:bg-sky-500"
        @click="toggleForm"
      >
        {{ showForm ? 'Cancel' : '+ Add expense' }}
      </button>
    </div>

    <div
      v-if="error"
      class="rounded-lg border border-red-900 bg-red-950/50 px-4 py-3 text-sm text-red-300"
    >
      {{ error }}
    </div>

    <section
      v-if="showForm"
      class="rounded-xl border border-zinc-800 bg-zinc-950 p-5"
    >
      <h3 class="text-base font-semibold text-zinc-100">
        {{ editingId ? 'Edit expense' : 'New expense' }}
      </h3>
      <form class="mt-4 space-y-4" @submit.prevent="submitExpense">
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="block text-sm">
            <span class="font-medium text-zinc-300">Paid by</span>
            <select
              v-model="form.person_id"
              class="mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100"
              required
            >
              <option v-for="p in persons" :key="p.id" :value="p.id">
                {{ p.name }}
              </option>
            </select>
          </label>

          <JalaliDateInput v-model="form.date" label="Date" required />
        </div>

        <div class="space-y-2">
          <span class="text-sm font-medium text-zinc-300">Shop</span>
          <div class="flex flex-wrap gap-3 text-zinc-300">
            <label class="inline-flex items-center gap-2 text-sm">
              <input v-model="form.useNewShop" type="radio" :value="false" />
              Existing
            </label>
            <label class="inline-flex items-center gap-2 text-sm">
              <input v-model="form.useNewShop" type="radio" :value="true" />
              New shop
            </label>
          </div>
          <select
            v-if="!form.useNewShop"
            v-model="form.shop_id"
            class="w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
            required
          >
            <option disabled value="">Select shop</option>
            <option v-for="s in shops" :key="s.id" :value="s.id">{{ s.name }}</option>
          </select>
          <input
            v-else
            v-model="form.newShopName"
            type="text"
            placeholder="Shop name"
            class="w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500"
            required
          />
        </div>

        <div>
          <div class="mb-2 flex items-center justify-between">
            <span class="text-sm font-medium text-zinc-300">Shares</span>
            <span
              class="text-xs"
              :class="sharesValid ? 'text-zinc-500' : 'text-red-400'"
            >
              Sum: {{ shareSum.toFixed(3) }}
            </span>
          </div>
          <div class="grid gap-2 sm:grid-cols-2">
            <label
              v-for="p in persons"
              :key="p.id"
              class="block text-sm"
            >
              <span class="font-medium text-zinc-300">{{ p.name }}</span>
              <input
                v-model="form.shares[p.id]"
                type="number"
                step="0.01"
                min="0"
                max="1"
                class="mt-1 w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-zinc-100"
                required
              />
            </label>
          </div>
          <p v-if="!sharesValid" class="mt-1 text-xs text-red-400">
            Shares must sum to 1
          </p>
        </div>

        <div>
          <div class="mb-2 flex items-center justify-between">
            <span class="text-sm font-medium text-zinc-300">Line items</span>
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
              <div class="relative sm:col-span-7">
                <input
                  v-model="item.description"
                  type="text"
                  placeholder="Item name (search after 3 chars)"
                  class="w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100 placeholder:text-zinc-500"
                  required
                  autocomplete="off"
                  @input="onItemNameInput(index)"
                  @focus="onItemNameInput(index)"
                  @blur="hideSuggestions(index)"
                />
                <ul
                  v-if="item.showSuggestions && item.suggestions.length"
                  class="absolute z-10 mt-1 max-h-40 w-full overflow-auto rounded border border-zinc-700 bg-zinc-950 shadow-lg"
                >
                  <li
                    v-for="s in item.suggestions"
                    :key="s.id"
                    class="cursor-pointer px-3 py-2 text-sm text-zinc-200 hover:bg-zinc-800"
                    @mousedown.prevent="pickSuggestion(index, s.name)"
                  >
                    {{ s.name }}
                  </li>
                </ul>
              </div>
              <input
                v-model="item.amount"
                type="number"
                step="0.01"
                min="0"
                placeholder="Amount"
                class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100 sm:col-span-3"
                required
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
          {{ saving ? 'Saving…' : editingId ? 'Update expense' : 'Save expense' }}
        </button>
      </form>
    </section>

    <section class="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div>
          <p class="text-sm text-zinc-400">
            {{ invoices.length }} records · {{ formatMoney(grandTotal) }} total
          </p>
        </div>
        <div class="grid gap-2 sm:grid-cols-3">
          <select
            v-model="filters.person_id"
            class="rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100"
            @change="loadData"
          >
            <option value="">All people</option>
            <option v-for="p in persons" :key="p.id" :value="p.id">{{ p.name }}</option>
          </select>
          <JalaliDateInput
            v-model="filters.from_date"
            :input-class="'w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500'"
            @change="loadData"
          />
          <JalaliDateInput
            v-model="filters.to_date"
            :input-class="'w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100 placeholder:text-zinc-500'"
            @change="loadData"
          />
        </div>
      </div>

      <div v-if="loading" class="py-12 text-center text-zinc-500">Loading…</div>
      <div v-else-if="invoices.length === 0" class="py-12 text-center text-zinc-500">
        No expenses yet. Add your first one above.
      </div>
      <ul v-else class="mt-4 divide-y divide-zinc-800">
        <li
          v-for="inv in invoices"
          :key="inv.id"
          class="flex flex-col gap-3 py-4 sm:flex-row sm:items-start sm:justify-between"
        >
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <span class="rounded-full bg-zinc-800 px-2.5 py-0.5 text-xs font-medium text-sky-300">
                Paid by {{ inv.person_name }}
              </span>
              <span class="text-sm font-medium text-zinc-100">{{ inv.shop_name }}</span>
              <span class="text-sm text-zinc-500">{{ toJalaliDisplay(inv.date) }}</span>
            </div>
            <ul class="mt-2 space-y-1 text-sm text-zinc-400">
              <li v-for="item in inv.items" :key="item.id">
                {{ item.description }} — {{ formatMoney(item.amount) }}
              </li>
            </ul>
            <ul
              v-if="inv.shares?.length"
              class="mt-2 flex flex-wrap gap-2 text-xs text-zinc-500"
            >
              <li
                v-for="s in inv.shares"
                :key="s.person_id"
                class="rounded bg-zinc-900 px-2 py-0.5"
              >
                {{ s.person_name }} {{ s.share }}
                ({{ formatMoney(inv.total * s.share) }})
              </li>
            </ul>
          </div>
          <div class="flex items-center gap-3 sm:flex-col sm:items-end">
            <span class="text-lg font-semibold text-white">{{ formatMoney(inv.total) }}</span>
            <div class="flex gap-3">
              <button
                class="text-sm text-sky-400 hover:text-sky-300"
                @click="startEdit(inv)"
              >
                Edit
              </button>
              <button
                class="text-sm text-red-400 hover:text-red-300"
                @click="deleteInvoice(inv.id)"
              >
                Delete
              </button>
            </div>
          </div>
        </li>
      </ul>
    </section>
  </div>
</template>
