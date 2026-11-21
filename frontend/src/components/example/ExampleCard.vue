<script setup lang="ts">
import { computed } from 'vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { ExampleItem } from '@/types/example'

interface Props {
  item: ExampleItem
}

const props = defineProps<Props>()

const emit = defineEmits<{
  edit: [item: ExampleItem]
  delete: [id: string]
}>()

const formattedDate = computed(() =>
  new Date(props.item.created_at).toLocaleDateString(undefined, {
    month: 'short',
    day: 'numeric',
    hour: '2-digit',
    minute: '2-digit',
  }),
)
</script>

<template>
  <article
    class="rounded-xl border border-slate-200 bg-white p-4 shadow-xs transition-shadow hover:shadow-sm"
  >
    <div class="flex flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
      <div class="min-w-0 flex-1">
        <h3 class="text-sm font-semibold text-slate-900">{{ item.name }}</h3>
        <p v-if="item.description" class="mt-1 text-sm text-slate-600">
          {{ item.description }}
        </p>
        <p class="mt-2 text-xs text-slate-400">Created {{ formattedDate }}</p>
      </div>

      <div class="flex shrink-0 items-center gap-2">
        <BaseButton variant="secondary" size="sm" @click="emit('edit', item)">Edit</BaseButton>
        <BaseButton variant="ghost" size="sm" @click="emit('delete', item.id)">Delete</BaseButton>
      </div>
    </div>
  </article>
</template>
