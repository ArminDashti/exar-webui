<script setup>
import { computed, onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { api } from '../api'
import { jalaliToGregorian, toJalaliDisplay } from '../utils/dates'
import { formatMoney } from '../utils/money'
import JalaliDateInput from '../components/JalaliDateInput.vue'

const router = useRouter()
const persons = ref([])
const expenses = ref([])
const loading = ref(true)
const error = ref('')
const editingId = ref(null)
const saving = ref(false)

const filters = reactive({
  person_id: '',
  from_date: '',
  to_date: '',
})

const editForm = reactive({
  person_id: 1,
  shop_name: '',
  shop_id: null,
  date: '',
  name: '',
  amount: '',
  armin_share: 0.5,
  ramin_share: 0.5,
  shopSuggestions: [],
  showShopSuggestions: false,
})

let shopTimer = null

const grandTotal = computed(() =>
  (expenses.value ?? []).reduce((sum, e) => sum + e.amount, 0)
)

function gregorianFilter() {
  const from = filters.from_date ? jalaliToGregorian(filters.from_date) : undefined
  const to = filters.to_date ? jalaliToGregorian(filters.to_date) : undefined
  if (filters.from_date && !from) throw new Error('Invalid from date')
  if (filters.to_date && !to) throw new Error('Invalid to date')
  return { from_date: from || undefined, to_date: to || undefined }
}

async function loadData() {
  loading.value = true
  error.value = ''
  try {
    const dates = gregorianFilter()
    const [p, list] = await Promise.all([
      api.getPersons(),
      api.getExpenses({
        person_id: filters.person_id || undefined,
        ...dates,
      }),
    ])
    persons.value = p ?? []
    expenses.value = list ?? []
  } catch (e) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

function shareOf(exp, personId) {
  return exp.shares?.find((s) => s.person_id === personId)?.share ?? 0
}

function setEditShare(person, value) {
  const n = parseFloat(value)
  if (Number.isNaN(n)) {
    if (person === 'armin') editForm.armin_share = value
    else editForm.ramin_share = value
    return
  }
  const clamped = Math.min(1, Math.max(0, n))
  if (person === 'armin') {
    editForm.armin_share = clamped
    editForm.ramin_share = Math.round((1 - clamped) * 1000) / 1000
  } else {
    editForm.ramin_share = clamped
    editForm.armin_share = Math.round((1 - clamped) * 1000) / 1000
  }
}

function startEdit(exp) {
  editingId.value = exp.id
  editForm.person_id = exp.person_id
  editForm.shop_name = exp.shop_name
  editForm.shop_id = exp.shop_id
  editForm.date = toJalaliDisplay(exp.date)
  editForm.name = exp.name
  editForm.amount = exp.amount
  editForm.armin_share = shareOf(exp, 1)
  editForm.ramin_share = shareOf(exp, 2)
  editForm.shopSuggestions = []
  editForm.showShopSuggestions = false
}

function cancelEdit() {
  editingId.value = null
}

function onShopInput() {
  editForm.shop_id = null
  clearTimeout(shopTimer)
  const q = editForm.shop_name.trim()
  if ([...q].length < 3) {
    editForm.shopSuggestions = []
    editForm.showShopSuggestions = false
    return
  }
  shopTimer = setTimeout(async () => {
    try {
      editForm.shopSuggestions = (await api.searchShops(q)) ?? []
      editForm.showShopSuggestions = true
    } catch {
      editForm.shopSuggestions = []
      editForm.showShopSuggestions = false
    }
  }, 250)
}

function pickShop(shop) {
  editForm.shop_name = shop.name
  editForm.shop_id = shop.id
  editForm.shopSuggestions = []
  editForm.showShopSuggestions = false
}

async function resolveShopId() {
  const name = editForm.shop_name.trim()
  if (!name) throw new Error('Shop is required')
  if (editForm.shop_id) return editForm.shop_id
  try {
    const shop = await api.createShop(name)
    return shop.id
  } catch (e) {
    if (/already exists/i.test(e.message)) {
      const all = await api.getShops()
      const found = (all ?? []).find((s) => s.name.toLowerCase() === name.toLowerCase())
      if (found) return found.id
    }
    throw e
  }
}

async function saveEdit() {
  saving.value = true
  error.value = ''
  try {
    const gregorianDate = jalaliToGregorian(editForm.date)
    if (!gregorianDate) throw new Error('Invalid date')
    const shopId = await resolveShopId()
    try {
      await api.createItem(editForm.name.trim())
    } catch (e) {
      if (!/already exists/i.test(e.message)) throw e
    }
    await api.updateExpense(editingId.value, {
      person_id: Number(editForm.person_id),
      shop_id: Number(shopId),
      date: gregorianDate,
      name: editForm.name.trim(),
      amount: parseFloat(editForm.amount),
      shares: [
        { person_id: 1, share: parseFloat(editForm.armin_share) || 0 },
        { person_id: 2, share: parseFloat(editForm.ramin_share) || 0 },
      ],
    })
    editingId.value = null
    await loadData()
  } catch (e) {
    error.value = e.message
  } finally {
    saving.value = false
  }
}

async function deleteExpense(id) {
  if (!confirm('Delete this expense?')) return
  try {
    await api.deleteExpense(id)
    await loadData()
  } catch (e) {
    error.value = e.message
  }
}

onMounted(loadData)
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-lg font-semibold text-white">Expenses</h2>
      <button
        class="rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold text-white hover:bg-sky-500"
        @click="router.push('/expenses/add')"
      >
        + Add
      </button>
    </div>

    <div
      v-if="error"
      class="rounded-lg border border-red-900 bg-red-950/50 px-4 py-3 text-sm text-red-300"
    >
      {{ error }}
    </div>

    <section class="rounded-xl border border-zinc-800 bg-zinc-950 p-5">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <p class="text-sm text-zinc-400">
          {{ expenses.length }} records · {{ formatMoney(grandTotal) }} total
        </p>
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
            :input-class="'w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100'"
            @change="loadData"
          />
          <JalaliDateInput
            v-model="filters.to_date"
            :input-class="'w-full rounded-lg border border-zinc-700 bg-zinc-900 px-3 py-2 text-sm text-zinc-100'"
            @change="loadData"
          />
        </div>
      </div>

      <div v-if="loading" class="py-12 text-center text-zinc-500">Loading…</div>
      <div v-else-if="expenses.length === 0" class="py-12 text-center text-zinc-500">
        No expenses yet.
      </div>
      <ul v-else class="mt-4 divide-y divide-zinc-800">
        <li v-for="exp in expenses" :key="exp.id" class="py-4">
          <div
            v-if="editingId === exp.id"
            class="space-y-3 rounded-lg border border-zinc-700 bg-zinc-900 p-4"
          >
            <div class="grid gap-3 sm:grid-cols-3">
              <label class="block text-sm">
                <span class="text-zinc-400">Paid by</span>
                <select
                  v-model="editForm.person_id"
                  class="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                >
                  <option v-for="p in persons" :key="p.id" :value="p.id">{{ p.name }}</option>
                </select>
              </label>
              <JalaliDateInput v-model="editForm.date" label="Date" required />
              <div class="relative block text-sm">
                <span class="text-zinc-400">Shop</span>
                <input
                  v-model="editForm.shop_name"
                  type="text"
                  class="mt-1 w-full rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                  @input="onShopInput"
                  @blur="editForm.showShopSuggestions = false"
                />
                <ul
                  v-if="editForm.showShopSuggestions && editForm.shopSuggestions.length"
                  class="absolute z-10 mt-1 max-h-40 w-full overflow-auto rounded border border-zinc-700 bg-zinc-950 shadow-lg"
                >
                  <li
                    v-for="s in editForm.shopSuggestions"
                    :key="s.id"
                    class="cursor-pointer px-3 py-2 text-sm hover:bg-zinc-800"
                    @mousedown.prevent="pickShop(s)"
                  >
                    {{ s.name }}
                  </li>
                </ul>
              </div>
            </div>
            <div class="grid gap-2 sm:grid-cols-4">
              <input
                v-model="editForm.name"
                type="text"
                placeholder="Item"
                class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
              />
              <input
                v-model="editForm.amount"
                type="number"
                step="1"
                min="0"
                placeholder="Amount"
                class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
              />
              <input
                :value="editForm.armin_share"
                type="number"
                step="0.01"
                min="0"
                max="1"
                placeholder="Armin share"
                class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                @input="setEditShare('armin', $event.target.value)"
              />
              <input
                :value="editForm.ramin_share"
                type="number"
                step="0.01"
                min="0"
                max="1"
                placeholder="Ramin share"
                class="rounded border border-zinc-700 bg-zinc-950 px-2 py-1.5 text-sm text-zinc-100"
                @input="setEditShare('ramin', $event.target.value)"
              />
            </div>
            <div class="flex gap-3">
              <button
                type="button"
                class="rounded-lg bg-sky-600 px-3 py-1.5 text-sm font-semibold text-white hover:bg-sky-500 disabled:opacity-60"
                :disabled="saving"
                @click="saveEdit"
              >
                {{ saving ? 'Saving…' : 'Save' }}
              </button>
              <button type="button" class="text-sm text-zinc-400 hover:text-zinc-200" @click="cancelEdit">
                Cancel
              </button>
            </div>
          </div>
          <div
            v-else
            class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between"
          >
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <span class="rounded-full bg-zinc-800 px-2.5 py-0.5 text-xs font-medium text-sky-300">
                  Paid by {{ exp.person_name }}
                </span>
                <span class="text-sm font-medium text-zinc-100">{{ exp.shop_name }}</span>
                <span class="text-sm text-zinc-500">{{ toJalaliDisplay(exp.date) }}</span>
              </div>
              <p class="mt-2 text-sm text-zinc-200">
                {{ exp.name }} — {{ formatMoney(exp.amount) }}
              </p>
              <ul class="mt-2 flex flex-wrap gap-2 text-xs text-zinc-500">
                <li
                  v-for="s in exp.shares"
                  :key="s.person_id"
                  class="rounded bg-zinc-900 px-2 py-0.5"
                >
                  {{ s.person_name }} {{ s.share }}
                  ({{ formatMoney(exp.amount * s.share) }})
                </li>
              </ul>
            </div>
            <div class="flex items-center gap-3 sm:flex-col sm:items-end">
              <span class="text-lg font-semibold text-white">{{ formatMoney(exp.amount) }}</span>
              <div class="flex gap-3">
                <button class="text-sm text-sky-400 hover:text-sky-300" @click="startEdit(exp)">
                  Edit
                </button>
                <button class="text-sm text-red-400 hover:text-red-300" @click="deleteExpense(exp.id)">
                  Delete
                </button>
              </div>
            </div>
          </div>
        </li>
      </ul>
    </section>
  </div>
</template>
