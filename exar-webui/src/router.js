import { createRouter, createWebHistory } from 'vue-router'
import ExpensesLayout from './views/ExpensesLayout.vue'
import ExpenseAddView from './views/ExpenseAddView.vue'
import ExpenseListView from './views/ExpenseListView.vue'
import ShopsView from './views/ShopsView.vue'
import ItemsView from './views/ItemsView.vue'
import StatsView from './views/StatsView.vue'
import LoginView from './views/LoginView.vue'
import { clearSession, isAuthenticated } from './auth'

const routes = [
  { path: '/login', name: 'login', component: LoginView, meta: { public: true } },
  { path: '/', redirect: '/expenses/list' },
  {
    path: '/expenses',
    component: ExpensesLayout,
    redirect: '/expenses/list',
    children: [
      { path: 'list', name: 'expenses-list', component: ExpenseListView },
      { path: 'add', name: 'expenses-add', component: ExpenseAddView },
    ],
  },
  { path: '/shops', name: 'shops', component: ShopsView },
  { path: '/items', name: 'items', component: ItemsView },
  { path: '/stats', name: 'stats', component: StatsView },
]

export const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes,
  linkActiveClass: 'text-sky-400',
  linkExactActiveClass: 'text-sky-400',
})

router.beforeEach((to) => {
  if (to.meta.public) {
    if (to.name === 'login' && isAuthenticated()) {
      return { path: '/expenses/list' }
    }
    return true
  }
  if (!isAuthenticated()) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }
  return true
})

export function logoutAndRedirect() {
  clearSession()
  return router.push({ name: 'login' })
}
