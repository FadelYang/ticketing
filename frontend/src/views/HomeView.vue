<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { storeToRefs } from 'pinia'
import { useExampleStore } from '@/stores/exampleStore'
import BaseBadge from '@/components/common/BaseBadge.vue'
import ExampleForm from '@/components/example/ExampleForm.vue'
import ExampleList from '@/components/example/ExampleList.vue'
import type { CreateExamplePayload, ExampleItem } from '@/types/example'

const exampleStore = useExampleStore()
const { items, loading, submitting, error, search, totalCount } = storeToRefs(exampleStore)

const editingItem = ref<ExampleItem | null>(null)

onMounted(() => {
  exampleStore.fetchItems()
})

async function handleFormSubmit(payload: CreateExamplePayload) {
  if (editingItem.value) {
    const ok = await exampleStore.updateItem(editingItem.value.id, payload)
    if (ok) {
      editingItem.value = null
    }
    return
  }

  await exampleStore.createItem(payload)
}

function handleDelete(id: string) {
  if (editingItem.value?.id === id) {
    editingItem.value = null
  }
  exampleStore.deleteItem(id)
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-center gap-2.5">
        <h2 class="text-lg font-semibold text-slate-900">Example Module</h2>
        <BaseBadge tone="blue">{{ totalCount }} items</BaseBadge>
      </div>

      <form class="flex items-center gap-2" @submit.prevent="exampleStore.fetchItems()">
        <input
          v-model="search"
          type="search"
          placeholder="Search examples..."
          class="w-full rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-sm text-slate-900 placeholder:text-slate-400 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 focus:outline-none sm:w-56"
        />
        <button
          type="submit"
          class="cursor-pointer rounded-lg border border-slate-300 bg-white px-3 py-1.5 text-xs font-medium text-slate-700 hover:bg-slate-50"
        >
          Search
        </button>
      </form>
    </div>

    <div
      v-if="error"
      class="rounded-xl border border-rose-200 bg-rose-50 px-4 py-3 text-sm text-rose-700"
    >
      {{ error }}
    </div>

    <ExampleForm
      :editing-item="editingItem"
      :submitting="submitting"
      @submit="handleFormSubmit"
      @cancel="editingItem = null"
    />

    <ExampleList
      :items="items"
      :loading="loading"
      @edit="editingItem = $event"
      @delete="handleDelete"
    />
  </div>
</template>
