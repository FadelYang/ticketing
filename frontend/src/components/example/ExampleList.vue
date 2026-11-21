<script setup lang="ts">
import EmptyState from '@/components/common/EmptyState.vue'
import ExampleCard from '@/components/example/ExampleCard.vue'
import type { ExampleItem } from '@/types/example'

interface Props {
  items: ExampleItem[]
  loading?: boolean
}

withDefaults(defineProps<Props>(), {
  loading: false,
})

const emit = defineEmits<{
  edit: [item: ExampleItem]
  delete: [id: string]
}>()
</script>

<template>
  <div v-if="loading" class="space-y-3">
    <div
      v-for="n in 3"
      :key="n"
      class="h-20 animate-pulse rounded-xl border border-slate-200 bg-white p-4"
    >
      <div class="h-4 w-1/3 rounded bg-slate-200" />
      <div class="mt-3 h-3 w-2/3 rounded bg-slate-100" />
    </div>
  </div>

  <EmptyState
    v-else-if="items.length === 0"
    title="No examples yet"
    description="Create your first example item above to verify the full stack."
  />

  <div v-else class="space-y-3">
    <ExampleCard
      v-for="item in items"
      :key="item.id"
      :item="item"
      @edit="emit('edit', $event)"
      @delete="emit('delete', $event)"
    />
  </div>
</template>
