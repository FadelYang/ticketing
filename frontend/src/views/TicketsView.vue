<script setup lang="ts">
import { computed, onMounted, watch } from 'vue'
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

watch(
  () => [filters.value.status, filters.value.priority, filters.value.assigneeId],
  () => {
    ticketStore.fetchTickets(1)
  },
)

watch(
  () => filters.value.search,
  () => {
    if (searchTimer) clearTimeout(searchTimer)
    searchTimer = setTimeout(() => {
      ticketStore.fetchTickets(1)
    }, 250)
  },
)

onMounted(() => {
  ticketStore.fetchUsers()
  ticketStore.fetchTickets(1)
})

function handleSelectTicket(_ticket: Ticket) {
  // Ticket detail view will be wired in the next step
}

function handleNewTicket() {
  router.push('/tickets/new')
}
</script>

<template>
  <DashboardLayout v-model:search-query="filters.search" :total-count="total">
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
