<script setup lang="ts">
import BaseIcon from '@/components/common/BaseIcon.vue'
import type { AuthUser, UserRole } from '@/types/auth'

interface Props {
  users: AuthUser[]
  loading: boolean
  updatingId: string | null
  currentUserId?: string | null
}

defineProps<Props>()

const emit = defineEmits<{
  'change-role': [userId: string, role: UserRole]
  reset: []
}>()

const roleOptions: { value: UserRole; label: string }[] = [
  { value: 'requester', label: 'Requester' },
  { value: 'agent', label: 'Agent' },
  { value: 'admin', label: 'Admin' },
]

const roleBadgeClass: Record<UserRole, string> = {
  admin: 'bg-primary-fixed text-on-primary-fixed-variant',
  agent: 'bg-secondary-fixed text-on-secondary-fixed',
  requester: 'bg-surface-container-high text-on-surface-variant',
}

function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return '?'
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return `${parts[0][0]}${parts[parts.length - 1][0]}`.toUpperCase()
}

function formatDate(iso: string): string {
  if (!iso) return '—'
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return iso
  return date.toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
  })
}
</script>

<template>
  <div class="bg-surface-container-lowest rounded-xl shadow-sm overflow-hidden flex flex-col">
    <div v-if="loading" class="p-space-2xl text-center text-outline flex items-center justify-center gap-2">
      <BaseIcon name="progress_activity" class="text-lg animate-spin" />
      <span class="font-body-md text-body-md">Loading users...</span>
    </div>

    <div
      v-else-if="users.length === 0"
      class="p-space-2xl flex flex-col items-center justify-center text-center gap-space-sm"
    >
      <BaseIcon name="group_off" class="text-4xl text-outline-variant" />
      <p class="font-headline-sm text-headline-sm text-on-surface">No users found</p>
      <p class="font-body-md text-body-md text-outline max-w-md">
        No user accounts match your current search or role filter.
      </p>
      <button
        type="button"
        class="mt-2 px-4 py-1.5 rounded-lg bg-primary text-on-primary font-label-md text-label-md cursor-pointer"
        @click="emit('reset')"
      >
        Reset Filters
      </button>
    </div>

    <div v-else class="overflow-x-auto">
      <table class="w-full text-left border-collapse min-w-[780px]">
        <thead>
          <tr
            class="border-b border-outline-variant/40 bg-surface-container-low text-on-surface-variant font-label-sm text-label-sm uppercase tracking-wider"
          >
            <th class="py-3 px-4 font-semibold">User</th>
            <th class="py-3 px-4 font-semibold">Email</th>
            <th class="py-3 px-4 font-semibold">Phone</th>
            <th class="py-3 px-4 font-semibold">Role</th>
            <th class="py-3 px-4 font-semibold">Joined</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-outline-variant/20 text-body-md font-body-md">
          <tr
            v-for="u in users"
            :key="u.id"
            class="hover:bg-surface-container-low/60 transition-colors"
          >
            <td class="py-3 px-4">
              <div class="flex items-center gap-space-sm">
                <div
                  class="w-8 h-8 rounded-full flex items-center justify-center font-label-sm text-label-sm font-bold shrink-0"
                  :class="roleBadgeClass[u.role] ?? roleBadgeClass.requester"
                >
                  {{ getInitials(u.name) }}
                </div>
                <div class="flex items-center gap-1.5 min-w-0">
                  <span class="font-semibold text-on-surface truncate">{{ u.name }}</span>
                  <span
                    v-if="u.id === currentUserId"
                    class="px-1.5 py-0.5 rounded bg-surface-container text-outline font-code-sm text-[10px] uppercase"
                  >
                    You
                  </span>
                </div>
              </div>
            </td>

            <td class="py-3 px-4 text-on-surface-variant font-code-sm text-code-sm">
              {{ u.email }}
            </td>

            <td class="py-3 px-4 text-on-surface-variant font-code-sm text-code-sm">
              {{ u.phonenumber }}
            </td>

            <td class="py-3 px-4">
              <div class="relative inline-flex items-center">
                <select
                  :value="u.role"
                  :disabled="updatingId === u.id || u.id === currentUserId"
                  aria-label="Change user role"
                  class="h-8 pl-2.5 pr-7 rounded-lg font-label-sm text-label-sm font-semibold appearance-none cursor-pointer focus:outline-none focus:ring-2 focus:ring-primary/40 transition-all disabled:opacity-60 disabled:cursor-not-allowed"
                  :class="roleBadgeClass[u.role] ?? roleBadgeClass.requester"
                  @change="emit('change-role', u.id, ($event.target as HTMLSelectElement).value as UserRole)"
                >
                  <option v-for="opt in roleOptions" :key="opt.value" :value="opt.value">
                    {{ opt.label }}
                  </option>
                </select>
                <BaseIcon
                  v-if="updatingId === u.id"
                  name="progress_activity"
                  class="absolute right-1.5 text-[14px] animate-spin pointer-events-none"
                />
                <BaseIcon
                  v-else
                  name="expand_more"
                  class="absolute right-1.5 text-[16px] pointer-events-none"
                />
              </div>
            </td>

            <td class="py-3 px-4 text-outline font-code-sm text-code-sm">
              {{ formatDate(u.created_at) }}
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
