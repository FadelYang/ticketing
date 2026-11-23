<script setup lang="ts">
import BaseIcon from '@/components/common/BaseIcon.vue'
import type { AuthUser } from '@/types/auth'
import type { TicketPriority } from '@/types/ticket'

interface Props {
  search: string
  priority: TicketPriority | 'all'
  assigneeId: string
  assignees: AuthUser[]
  showAssigneeFilter?: boolean
}

withDefaults(defineProps<Props>(), {
  showAssigneeFilter: true,
})

const emit = defineEmits<{
  'update:search': [value: string]
  'update:priority': [value: TicketPriority | 'all']
  'update:assigneeId': [value: string]
  reset: []
}>()
</script>

<template>
  <div
    class="bg-surface-container-lowest rounded-xl p-space-md shadow-sm flex flex-col md:flex-row md:items-center justify-between gap-space-md"
  >
    <!-- Search Field -->
    <div class="relative flex-1 max-w-lg">
      <BaseIcon
        name="search"
        class="absolute left-3 top-1/2 -translate-y-1/2 text-outline text-[20px] pointer-events-none"
      />
      <input
        :value="search"
        type="text"
        placeholder="Search tickets by title, issue ID, or summary..."
        class="w-full pl-11 pr-8 py-2 bg-surface-container-low rounded-lg text-on-surface placeholder:text-outline font-body-sm text-body-sm focus:outline-none focus:bg-surface-container-lowest transition-all"
        @input="emit('update:search', ($event.target as HTMLInputElement).value)"
      />
      <button
        v-if="search"
        type="button"
        aria-label="Clear search"
        class="absolute right-2.5 top-1/2 -translate-y-1/2 text-outline hover:text-on-surface p-1 cursor-pointer"
        @click="emit('update:search', '')"
      >
        <BaseIcon name="close" class="text-[16px]" />
      </button>
    </div>

    <!-- Dropdown Filters Group -->
    <div class="flex flex-wrap items-center gap-space-sm">
      <div class="relative">
        <select
          :value="priority"
          class="appearance-none bg-surface-container-low text-on-surface font-label-md text-label-md pl-3 pr-8 py-2 rounded-lg cursor-pointer focus:outline-none hover:bg-surface-container-high transition-colors"
          @change="emit('update:priority', ($event.target as HTMLSelectElement).value as TicketPriority | 'all')"
        >
          <option value="all">All Priorities</option>
          <option value="urgent">Urgent</option>
          <option value="high">High</option>
          <option value="medium">Medium</option>
          <option value="low">Low</option>
        </select>
        <BaseIcon
          name="expand_more"
          class="absolute right-2 top-1/2 -translate-y-1/2 text-outline text-[18px] pointer-events-none"
        />
      </div>

      <div v-if="showAssigneeFilter" class="relative">
        <select
          :value="assigneeId"
          class="appearance-none bg-surface-container-low text-on-surface font-label-md text-label-md pl-3 pr-8 py-2 rounded-lg cursor-pointer focus:outline-none hover:bg-surface-container-high transition-colors"
          @change="emit('update:assigneeId', ($event.target as HTMLSelectElement).value)"
        >
          <option value="all">All Assignees</option>
          <option v-for="agent in assignees" :key="agent.id" :value="agent.id">
            {{ agent.name }}
          </option>
        </select>
        <BaseIcon
          name="expand_more"
          class="absolute right-2 top-1/2 -translate-y-1/2 text-outline text-[18px] pointer-events-none"
        />
      </div>

      <button
        type="button"
        class="px-3 py-2 rounded-lg bg-surface-container-low hover:bg-surface-container-high text-on-surface-variant font-label-md text-label-md flex items-center gap-1.5 transition-colors cursor-pointer"
        @click="emit('reset')"
      >
        <BaseIcon name="filter_alt_off" class="text-[16px]" />
        <span>Clear</span>
      </button>
    </div>
  </div>
</template>
