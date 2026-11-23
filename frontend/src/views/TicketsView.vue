<script setup lang="ts">
import { computed, onMounted, onUnmounted, watch } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/authStore'
import { useTicketStore } from '@/stores/ticketStore'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import TicketStatusTabs from '@/components/tickets/TicketStatusTabs.vue'
import TicketFilterBar from '@/components/tickets/TicketFilterBar.vue'
import TicketTable from '@/components/tickets/TicketTable.vue'
import BaseAlert from '@/components/common/BaseAlert.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'
import type { Ticket } from '@/types/ticket'

const router = useRouter()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)

const ticketStore = useTicketStore()
const {
  tickets,
  lateTickets,
  assignees,
  userMap,
  total,
  page,
  totalPages,
  loading,
  error,
  filters,
  statusCounts,
} = storeToRefs(ticketStore)

const isRequester = computed(() => user.value?.role === 'requester')
const canCreateTicket = computed(() => user.value?.role === 'requester')

const pageTitle = computed(() => (isRequester.value ? 'My Tickets' : 'Tickets'))

let searchTimer: ReturnType<typeof setTimeout> | null = null
let latePollTimer: ReturnType<typeof setInterval> | null = null

watch(
  () => filters.value.status,
  () => {
    ticketStore.fetchTickets(1)
  },
)

watch(
  () => [filters.value.priority, filters.value.assigneeId],
  () => {
    ticketStore.fetchTickets(1)
    ticketStore.fetchStatusCounts()
  },
)

watch(
  () => filters.value.search,
  () => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      ticketStore.fetchTickets(1)
      ticketStore.fetchStatusCounts()
    }, 250)
  },
)

onMounted(() => {
  ticketStore.fetchUsers()
  ticketStore.fetchTickets(1)
  ticketStore.fetchStatusCounts()
  ticketStore.fetchLateTickets()
  latePollTimer = setInterval(() => {
    void ticketStore.fetchLateTickets()
  }, 60_000)
})

onUnmounted(() => {
  if (latePollTimer) {
    clearInterval(latePollTimer)
    latePollTimer = null
  }
})

function handleSelectTicket(ticket: Ticket) {
  router.push({ name: 'ticket-detail', params: { id: ticket.id } })
}

function handleNewTicket() {
  router.push('/tickets/new')
}
</script>

<template>
  <DashboardLayout v-model:search-query="filters.search" :total-count="statusCounts.all">
    <div class="flex flex-col w-full gap-space-lg">
      <!-- Operational Header -->
      <div class="flex flex-col md:flex-row md:items-center justify-between gap-space-md">
        <div class="flex flex-col">
          <div class="flex items-center gap-space-sm mt-2">
            <span class="font-headline-lg text-headline-lg text-on-surface font-bold tracking-tight">
              {{ pageTitle }}
            </span>
          </div>
          <p class="font-body-md text-body-md text-outline mt-0.5">
            Showing
            <span class="font-semibold text-on-surface">{{ tickets.length }}</span>
            of {{ total }} total active dispatch records
          </p>
        </div>

        <div v-if="canCreateTicket" class="flex items-center gap-space-md">
          <button
            type="button"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary hover:bg-primary/90 text-on-primary font-label-md text-label-md font-semibold shadow-sm transition-all focus:outline-none focus:ring-2 focus:ring-primary/40 cursor-pointer"
            @click="handleNewTicket"
          >
            <BaseIcon name="add" class="text-[20px]" />
            <span>New Ticket</span>
          </button>
        </div>
      </div>

      <BaseAlert v-if="error" variant="error" :message="error" />

      <!-- Segmented Status Filter -->
      <TicketStatusTabs v-model="filters.status" :counts="statusCounts" />

      <!-- Search & Dropdown Filter Bar -->
      <TicketFilterBar
        v-model:search="filters.search"
        v-model:priority="filters.priority"
        v-model:assignee-id="filters.assigneeId"
        :assignees="assignees"
        :show-assignee-filter="!isRequester"
        @reset="ticketStore.resetFilters()"
      />

      <!-- Late Urgent Tickets Alert Above Table -->
      <div
        v-if="lateTickets.length > 0"
        role="alert"
        class="rounded-xl bg-error-container/70 border border-error/30 p-space-md flex flex-col gap-space-xs text-on-error-container shadow-xs"
      >
        <div class="flex items-center gap-2 font-label-md text-label-md font-bold">
          <BaseIcon name="warning" class="text-[20px] text-error shrink-0" />
          <span>
            {{ lateTickets.length }} urgent {{ lateTickets.length === 1 ? 'ticket has' : 'tickets have' }}
            not been handled within 3 hours
          </span>
        </div>
        <div class="flex flex-wrap gap-2 pl-7">
          <RouterLink
            v-for="late in lateTickets"
            :key="late.id"
            :to="{ name: 'ticket-detail', params: { id: late.id } }"
            class="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-lg bg-surface-container-lowest/90 hover:bg-surface-container-lowest text-on-surface font-body-sm text-body-sm font-medium border border-error/20 transition-colors"
          >
            <span class="w-1.5 h-1.5 rounded-full bg-error shrink-0" />
            <span class="truncate max-w-[260px]">{{ late.title }}</span>
          </RouterLink>
        </div>
      </div>

      <!-- Master Tickets Table -->
      <TicketTable
        :tickets="tickets"
        :user-map="userMap"
        :loading="loading"
        :page="page"
        :total-pages="totalPages"
        @select="handleSelectTicket"
        @reset="ticketStore.resetFilters()"
        @change-page="ticketStore.fetchTickets($event)"
      />
    </div>
  </DashboardLayout>
</template>
