<script setup lang="ts">
import BaseIcon from '@/components/common/BaseIcon.vue'
import type { AuthUser } from '@/types/auth'
import type { Ticket, TicketPriority, TicketStatus } from '@/types/ticket'

interface Props {
  tickets: Ticket[]
  userMap: Map<string, AuthUser>
  loading: boolean
  page: number
  totalPages: number
}

const props = defineProps<Props>()

const emit = defineEmits<{
  select: [ticket: Ticket]
  reset: []
  'change-page': [page: number]
}>()

const statusConfig: Record<TicketStatus, { label: string; badgeClass: string }> = {
  open: {
    label: 'Open',
    badgeClass: 'bg-surface-container-high text-on-surface-variant',
  },
  in_progress: {
    label: 'In Progress',
    badgeClass: 'bg-primary-fixed text-on-primary-fixed-variant',
  },
  resolved: {
    label: 'Resolved',
    badgeClass: 'bg-surface-container text-on-surface',
  },
  closed: {
    label: 'Closed',
    badgeClass: 'bg-surface-container-high text-outline',
  },
}

const priorityConfig: Record<TicketPriority, { label: string; badgeClass: string }> = {
  urgent: {
    label: 'Urgent',
    badgeClass: 'bg-error-container text-on-error-container font-semibold',
  },
  high: {
    label: 'High',
    badgeClass: 'bg-surface-container-high text-on-surface font-semibold',
  },
  medium: {
    label: 'Medium',
    badgeClass: 'bg-surface-container text-on-surface-variant font-medium',
  },
  low: {
    label: 'Low',
    badgeClass: 'bg-surface-container text-outline font-medium',
  },
}

function getAssignee(ticket: Ticket): { name: string; initials: string; assigned: boolean } {
  if (!ticket.assignee_id) {
    return { name: 'Unassigned', initials: '--', assigned: false }
  }
  const user = props.userMap.get(ticket.assignee_id)
  if (!user) {
    return { name: ticket.assignee_id.slice(0, 8), initials: 'AG', assigned: true }
  }
  const parts = user.name.trim().split(/\s+/)
  const initials =
    parts.length >= 2
      ? `${parts[0][0]}${parts[1][0]}`.toUpperCase()
      : user.name.slice(0, 2).toUpperCase()
  return { name: user.name, initials, assigned: true }
}

function formatDeadline(iso: string): string {
  if (!iso) return '—'
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return iso
  return date.toLocaleString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  })
}
</script>

<template>
  <div class="bg-surface-container-lowest rounded-xl shadow-sm overflow-hidden flex flex-col">
    <div v-if="loading" class="p-space-2xl flex items-center justify-center gap-2 text-outline">
      <BaseIcon name="progress_activity" class="text-lg animate-spin" />
      <span class="font-body-md text-body-md">Loading tickets...</span>
    </div>

    <div v-else-if="tickets.length > 0" class="overflow-x-auto w-full">
      <table class="w-full text-left border-collapse">
        <thead>
          <tr
            class="bg-surface-container-low/70 text-outline font-label-sm text-label-sm uppercase tracking-wider select-none"
          >
            <th class="py-2 px-4 font-semibold min-w-[280px]">Subject &amp; Overview</th>
            <th class="py-2 px-4 font-semibold w-36">Status</th>
            <th class="py-2 px-4 font-semibold w-32">Priority</th>
            <th class="py-2 px-4 font-semibold w-48">Assignee</th>
            <th class="py-2 px-4 font-semibold w-36">Deadline</th>
            <th class="py-2 px-4 font-semibold text-right w-20">Action</th>
          </tr>
        </thead>
        <tbody class="text-on-surface font-body-sm text-body-sm divide-y divide-outline-variant/20">
          <tr
            v-for="ticket in tickets"
            :key="ticket.id"
            class="group transition-colors hover:bg-surface-container-low/40 cursor-pointer"
            @click="emit('select', ticket)"
          >
            <td class="py-2 px-4">
              <div class="flex flex-col">
                <span
                  class="font-headline-sm text-headline-sm font-semibold group-hover:text-primary transition-colors leading-tight"
                  :class="
                    ticket.status === 'closed'
                      ? 'text-on-surface-variant line-through'
                      : 'text-on-surface'
                  "
                >
                  {{ ticket.title }}
                </span>
                <span class="text-outline text-body-sm truncate max-w-md">
                  {{ ticket.description }}
                </span>
              </div>
            </td>

            <td class="py-2 px-4">
              <span
                class="inline-flex items-center px-2 py-0.5 rounded-full text-label-sm font-label-sm font-medium whitespace-nowrap shrink-0"
                :class="statusConfig[ticket.status]?.badgeClass"
              >
                {{ statusConfig[ticket.status]?.label ?? ticket.status }}
              </span>
            </td>

            <td class="py-2 px-4">
              <span
                class="inline-flex items-center px-2 py-0.5 rounded-full text-label-sm font-label-sm"
                :class="priorityConfig[ticket.priority]?.badgeClass"
              >
                {{ priorityConfig[ticket.priority]?.label ?? ticket.priority }}
              </span>
            </td>

            <td class="py-2 px-4">
              <div class="flex items-center gap-2">
                <div
                  class="w-5 h-5 rounded-full flex items-center justify-center font-bold text-[10px]"
                  :class="
                    getAssignee(ticket).assigned
                      ? 'bg-primary text-on-primary'
                      : 'bg-surface-container-high text-outline font-medium'
                  "
                >
                  {{ getAssignee(ticket).initials }}
                </div>
                <span
                  class="font-medium"
                  :class="getAssignee(ticket).assigned ? 'text-on-surface' : 'text-outline italic'"
                >
                  {{ getAssignee(ticket).name }}
                </span>
              </div>
            </td>

            <td class="py-2 px-4 text-outline font-code-sm text-code-sm">
              {{ formatDeadline(ticket.deadline_at) }}
            </td>

            <td class="py-2 px-4 text-right">
              <button
                type="button"
                title="Quick View"
                class="p-0.5 rounded hover:bg-surface-container-high text-outline hover:text-on-surface transition-colors cursor-pointer"
                @click.stop="emit('select', ticket)"
              >
                <BaseIcon name="chevron_right" class="text-[18px]" />
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Empty State -->
    <div v-else class="flex flex-col items-center justify-center py-16 px-4 text-center">
      <div
        class="w-12 h-12 rounded-full bg-surface-container-high flex items-center justify-center text-outline mb-space-sm"
      >
        <BaseIcon name="search_off" class="text-[28px]" />
      </div>
      <span class="font-headline-md text-headline-md font-semibold text-on-surface">
        No tickets match criteria
      </span>
      <p class="font-body-md text-body-md text-outline max-w-sm mt-1">
        Try modifying your keywords or clearing the active priority, status, or assignee dropdown
        filters.
      </p>
      <button
        type="button"
        class="mt-4 px-space-md py-1.5 rounded-lg bg-surface-container-high text-primary hover:bg-surface-container-highest font-label-md text-label-md transition-colors cursor-pointer"
        @click="emit('reset')"
      >
        Reset All Filters
      </button>
    </div>

    <!-- Pagination Footer -->
    <div class="flex items-center justify-between px-4 py-3 bg-surface-container-low/50">
      <span class="font-label-sm text-label-sm text-outline">
        Live ticket feed • <span class="text-primary font-medium">99.98% SLA Compliance</span>
      </span>
      <div class="flex items-center gap-2">
        <button
          type="button"
          :disabled="page <= 1"
          class="p-1 rounded text-outline hover:text-on-surface hover:bg-surface-container-high disabled:opacity-40 cursor-pointer disabled:cursor-not-allowed"
          @click="emit('change-page', page - 1)"
        >
          <BaseIcon name="chevron_left" class="text-[18px]" />
        </button>
        <span class="font-code-sm text-code-sm px-2 text-on-surface-variant">
          Page {{ page }} of {{ totalPages }}
        </span>
        <button
          type="button"
          :disabled="page >= totalPages"
          class="p-1 rounded text-outline hover:text-on-surface hover:bg-surface-container-high disabled:opacity-40 cursor-pointer disabled:cursor-not-allowed"
          @click="emit('change-page', page + 1)"
        >
          <BaseIcon name="chevron_right" class="text-[18px]" />
        </button>
      </div>
    </div>
  </div>
</template>
