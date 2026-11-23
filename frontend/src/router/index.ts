import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/authStore'
import LoginView from '@/views/LoginView.vue'
import RegisterView from '@/views/RegisterView.vue'
import TicketsView from '@/views/TicketsView.vue'
import CreateTicketView from '@/views/CreateTicketView.vue'
import TicketDetailView from '@/views/TicketDetailView.vue'
import TeamUsersView from '@/views/TeamUsersView.vue'
import CreateUserView from '@/views/CreateUserView.vue'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      redirect: '/tickets',
    },
    {
      path: '/tickets',
      name: 'tickets',
      component: TicketsView,
      meta: { hideNavbar: true, requiresAuth: true },
    },
    {
      path: '/tickets/new',
      name: 'tickets-new',
      component: CreateTicketView,
      meta: { hideNavbar: true, requiresAuth: true, requesterOnly: true },
    },
    {
      path: '/tickets/:id',
      name: 'ticket-detail',
      component: TicketDetailView,
      meta: { hideNavbar: true, requiresAuth: true },
    },
    {
      path: '/team',
      name: 'team',
      component: TeamUsersView,
      meta: { hideNavbar: true, requiresAuth: true, adminOnly: true },
    },
    {
      path: '/team/new',
      name: 'team-new',
      component: CreateUserView,
      meta: { hideNavbar: true, requiresAuth: true, adminOnly: true },
    },
    {
      path: '/login',
      name: 'login',
      component: LoginView,
      meta: { hideNavbar: true, guestOnly: true },
    },
    {
      path: '/register',
      name: 'register',
      component: RegisterView,
      meta: { hideNavbar: true, guestOnly: true },
    },
  ],
})

router.beforeEach((to) => {
  const authStore = useAuthStore()

  if (to.meta.requiresAuth && !authStore.isAuthenticated) {
    return { name: 'login', query: { redirect: to.fullPath } }
  }

  if (to.meta.requesterOnly && authStore.user?.role !== 'requester') {
    return { name: 'tickets' }
  }

  if (to.meta.adminOnly && authStore.user?.role !== 'admin') {
    return { name: 'tickets' }
  }

  if (to.meta.guestOnly && authStore.isAuthenticated) {
    return { name: 'tickets' }
  }

  return true
})

export default router
