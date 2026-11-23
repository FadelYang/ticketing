<script setup lang="ts">
import BaseIcon from '@/components/common/BaseIcon.vue'
import { useTicketSocket } from '@/composables/useTicketSocket'

const { notifications, dismissNotification } = useTicketSocket()
</script>

<template>
  <Teleport to="body">
    <div
      v-if="notifications.length > 0"
      aria-live="polite"
      aria-label="Ticket notifications"
      class="fixed top-20 right-4 z-[60] flex flex-col gap-space-sm w-[calc(100vw-2rem)] max-w-sm pointer-events-none"
    >
      <div
        v-for="item in notifications"
        :key="item.id"
        role="alert"
        class="pointer-events-auto rounded-xl bg-surface-container-lowest border border-primary/30 shadow-lg p-space-md flex items-start gap-space-sm transition-all"
      >
        <div
          class="w-9 h-9 rounded-lg bg-primary-fixed text-on-primary-fixed-variant flex items-center justify-center shrink-0"
        >
          <BaseIcon name="notifications_active" class="text-[20px]" />
        </div>

        <div class="flex-1 min-w-0">
          <div class="flex items-center justify-between gap-2">
            <span class="font-label-sm text-label-sm uppercase tracking-wider font-bold text-primary">
              New Ticket
            </span>
            <span
              class="px-2 py-0.5 rounded-full bg-surface-container-high text-on-surface-variant font-label-sm text-label-sm capitalize"
            >
              {{ item.ticket.priority }}
            </span>
          </div>

          <p class="font-body-md text-body-md font-semibold text-on-surface truncate mt-0.5">
            {{ item.ticket.title }}
          </p>

          <div class="flex items-center justify-between gap-2 mt-1.5">
            <span class="font-code-sm text-code-sm text-outline truncate">
              {{ item.ticket.category }}
            </span>
            <RouterLink
              :to="{ name: 'ticket-detail', params: { id: item.ticket.id } }"
              class="font-label-sm text-label-sm font-semibold text-primary hover:underline shrink-0"
              @click="dismissNotification(item.id)"
            >
              View ticket
            </RouterLink>
          </div>
        </div>

        <button
          type="button"
          class="p-1 rounded-lg text-outline hover:text-on-surface hover:bg-surface-container transition-colors cursor-pointer shrink-0"
          aria-label="Dismiss notification"
          @click="dismissNotification(item.id)"
        >
          <BaseIcon name="close" class="text-[18px]" />
        </button>
      </div>
    </div>
  </Teleport>
</template>
