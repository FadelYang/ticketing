<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ApiError } from '@/services/api'
import { userService } from '@/services/userService'
import { useAuthStore } from '@/stores/authStore'
import { useTicketStore } from '@/stores/ticketStore'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import UserTable from '@/components/team/UserTable.vue'
import BaseAlert from '@/components/common/BaseAlert.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'
import type { AuthUser, UserRole } from '@/types/auth'

const router = useRouter()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)

const ticketStore = useTicketStore()
const { statusCounts } = storeToRefs(ticketStore)

const users = ref<AuthUser[]>([])
const loading = ref(true)
const updatingId = ref<string | null>(null)
const error = ref<string | null>(null)
const actionSuccess = ref<string | null>(null)

const search = ref('')
const roleFilter = ref<UserRole | 'all'>('all')

const roleTabs: { key: UserRole | 'all'; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'admin', label: 'Admin' },
  { key: 'agent', label: 'Agent' },
  { key: 'requester', label: 'Requester' },
]

const roleCounts = computed(() => {
  const counts: Record<UserRole | 'all', number> = {
    all: users.value.length,
    admin: 0,
    agent: 0,
    requester: 0,
  }
  for (const u of users.value) {
    if (u.role in counts) {
      counts[u.role]++
    }
  }
  return counts
})

const filteredUsers = computed(() => {
  const q = search.value.trim().toLowerCase()
  return users.value.filter((u) => {
    if (roleFilter.value !== 'all' && u.role !== roleFilter.value) {
      return false
    }
    if (!q) return true
    return (
      u.name.toLowerCase().includes(q) ||
      u.email.toLowerCase().includes(q) ||
      u.phonenumber.toLowerCase().includes(q) ||
      u.role.toLowerCase().includes(q)
    )
  })
})

async function fetchUsers() {
  loading.value = true
  error.value = null
  try {
    users.value = await userService.list()
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Failed to load users.'
  } finally {
    loading.value = false
  }
}

async function handleRoleChange(userId: string, nextRole: UserRole) {
  updatingId.value = userId
  error.value = null
  actionSuccess.value = null
  try {
    const updated = await userService.updateRole(userId, nextRole)
    users.value = users.value.map((u) => (u.id === userId ? updated : u))
    actionSuccess.value = `Updated ${updated.name}'s role to ${updated.role}.`
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Failed to update user role.'
  } finally {
    updatingId.value = null
  }
}

function resetFilters() {
  search.value = ''
  roleFilter.value = 'all'
}

onMounted(() => {
  void fetchUsers()
  void ticketStore.fetchStatusCounts()
})
</script>

<template>
  <DashboardLayout v-model:search-query="search" :total-count="statusCounts.all">
    <div class="flex flex-col w-full gap-space-lg">
      <!-- Operational Header -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-space-md">
        <div class="flex flex-col">
          <div class="flex items-center gap-space-sm mt-2">
            <span class="font-headline-lg text-headline-lg text-on-surface font-bold tracking-tight">
              Team &amp; Users
            </span>
          </div>
          <p class="font-body-md text-body-md text-outline mt-0.5">
            Showing
            <span class="font-semibold text-on-surface">{{ filteredUsers.length }}</span>
            of {{ users.length }} total user accounts
          </p>
        </div>

        <div class="flex items-center gap-space-md">
          <button
            type="button"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-on-primary font-label-md text-label-md font-semibold shadow-sm transition-all focus:outline-none focus:ring-2 focus:ring-primary/40 cursor-pointer"
            @click="router.push('/team/new')"
          >
            <BaseIcon name="person_add" class="text-[20px]" />
            <span>New User</span>
          </button>
        </div>
      </div>

      <BaseAlert v-if="error" variant="error" :message="error" />
      <BaseAlert v-if="actionSuccess" variant="success" :message="actionSuccess" />

      <!-- Segmented Role Filter -->
      <div class="w-full bg-surface-container-low p-1 rounded-xl shadow-sm flex items-stretch overflow-x-auto">
        <div class="grid grid-cols-4 w-full min-w-[480px] gap-1 select-none">
          <button
            v-for="tab in roleTabs"
            :key="tab.key"
            type="button"
            class="flex items-center justify-center gap-2 py-2 px-3 rounded-lg font-label-md text-label-md transition-all duration-150 cursor-pointer"
            :class="
              roleFilter === tab.key
                ? 'bg-surface-container-lowest text-primary font-bold shadow-sm'
                : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high/60 font-medium'
            "
            @click="roleFilter = tab.key"
          >
            <span>{{ tab.label }}</span>
            <span
              class="font-code-sm text-code-sm px-1.5 py-0.5 rounded-full bg-surface-container text-on-surface-variant"
            >
              {{ roleCounts[tab.key] }}
            </span>
          </button>
        </div>
      </div>

      <!-- Search Filter Bar -->
      <div class="flex flex-col sm:flex-row items-stretch sm:items-center gap-space-md">
        <div class="relative flex-1">
          <BaseIcon
            name="search"
            class="absolute left-3 top-1/2 -translate-y-1/2 text-outline text-[18px] pointer-events-none"
          />
          <input
            v-model="search"
            type="text"
            placeholder="Search users by name, email, phone, or role..."
            class="w-full h-10 pl-9 pr-3 rounded-lg bg-surface-container-lowest border border-outline-variant/50 text-on-surface placeholder:text-outline font-body-md text-body-md focus:outline-none focus:border-primary transition-all"
          />
        </div>
        <button
          v-if="search || roleFilter !== 'all'"
          type="button"
          class="h-10 px-4 rounded-lg bg-surface-container hover:bg-surface-container-high text-on-surface font-label-md text-label-md font-semibold transition-colors cursor-pointer shrink-0"
          @click="resetFilters"
        >
          Reset
        </button>
      </div>

      <!-- Master Users Table -->
      <UserTable
        :users="filteredUsers"
        :loading="loading"
        :updating-id="updatingId"
        :current-user-id="user?.id"
        @change-role="handleRoleChange"
        @reset="resetFilters"
      />
    </div>
  </DashboardLayout>
</template>
