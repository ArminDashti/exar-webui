import { createRouter, createWebHistory } from 'vue-router'
import ExpensesView from './views/ExpensesView.vue'
import ShopsView from './views/ShopsView.vue'
import ItemsView from './views/ItemsView.vue'
import StatsView from './views/StatsView.vue'

const routes = [
  { path: '/', redirect: '/expenses' },
  { path: '/expenses', name: 'expenses', component: ExpensesView },
  { path: '/shops', name: 'shops', component: ShopsView },
  { path: '/items', name: 'items', component: ItemsView },
  { path: '/stats', name: 'stats', component: StatsView },
]

export const router = createRouter({
  history: createWebHistory(),
  routes,
  linkActiveClass: 'text-sky-400',
  linkExactActiveClass: 'text-sky-400',
})
