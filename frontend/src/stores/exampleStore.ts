import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { exampleService } from '@/services/exampleService'
import type { CreateExamplePayload, ExampleItem, UpdateExamplePayload } from '@/types/example'

export const useExampleStore = defineStore('examples', () => {
  const items = ref<ExampleItem[]>([])
  const loading = ref(false)
  const submitting = ref(false)
  const error = ref<string | null>(null)
  const search = ref('')

  const totalCount = computed(() => items.value.length)

  async function fetchItems() {
    loading.value = true
    error.value = null
    try {
      items.value = await exampleService.list(search.value)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to load examples'
    } finally {
      loading.value = false
    }
  }

  async function createItem(payload: CreateExamplePayload): Promise<boolean> {
    submitting.value = true
    error.value = null
    try {
      const created = await exampleService.create(payload)
      items.value.unshift(created)
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to create example'
      return false
    } finally {
      submitting.value = false
    }
  }

  async function updateItem(id: string, payload: UpdateExamplePayload): Promise<boolean> {
    submitting.value = true
    error.value = null
    try {
      const updated = await exampleService.update(id, payload)
      const idx = items.value.findIndex((item) => item.id === id)
      if (idx !== -1) {
        items.value[idx] = updated
      }
      return true
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to update example'
      return false
    } finally {
      submitting.value = false
    }
  }

  async function deleteItem(id: string) {
    error.value = null
    try {
      await exampleService.remove(id)
      items.value = items.value.filter((item) => item.id !== id)
    } catch (err) {
      error.value = err instanceof Error ? err.message : 'Failed to delete example'
    }
  }

  return {
    items,
    loading,
    submitting,
    error,
    search,
    totalCount,
    fetchItems,
    createItem,
    updateItem,
    deleteItem,
  }
})
