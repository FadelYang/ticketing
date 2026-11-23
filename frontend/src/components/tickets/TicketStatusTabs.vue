<script setup lang="ts">
import type { TicketStatus } from '@/types/ticket'

interface Props {
  modelValue: TicketStatus | 'all'
  counts: Record<TicketStatus | 'all', number>
}

defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: TicketStatus | 'all']
}>()

const tabs: { key: TicketStatus | 'all'; label: string }[] = [
  { key: 'all', label: 'All' },
  { key: 'open', label: 'Open' },
  { key: 'in_progress', label: 'In Progress' },
  { key: 'resolved', label: 'Resolved' },
  { key: 'closed', label: 'Closed' },
]
</script>

<template>
  <div class="w-full bg-surface-container-low p-1 rounded-xl shadow-sm flex items-stretch overflow-x-auto">
    <div class="grid grid-cols-5 w-full min-w-[620px] gap-1 select-none">
      <button
        v-for="tab in tabs"
        :key="tab.key"
        type="button"
        class="flex items-center justify-center gap-2 py-2 px-3 rounded-lg font-label-md text-label-md transition-all duration-150 cursor-pointer"
        :class="
          modelValue === tab.key
            ? 'bg-surface-container-lowest text-primary font-bold shadow-sm'
            : 'text-on-surface-variant hover:text-on-surface hover:bg-surface-container-high/60 font-medium'
        "
        @click="emit('update:modelValue', tab.key)"
      >
        <span>{{ tab.label }}</span>
        <span
          class="font-code-sm text-code-sm px-1.5 py-0.5 rounded-full bg-surface-container text-on-surface-variant"
          :class="modelValue === tab.key ? 'font-medium' : ''"
        >
          {{ counts[tab.key] }}
        </span>
      </button>
    </div>
  </div>
</template>
