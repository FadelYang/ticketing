<script setup lang="ts">
import { reactive, watch } from 'vue'
import BaseButton from '@/components/common/BaseButton.vue'
import type { CreateExamplePayload, ExampleItem } from '@/types/example'

interface Props {
  editingItem?: ExampleItem | null
  submitting?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  editingItem: null,
  submitting: false,
})

const emit = defineEmits<{
  submit: [payload: CreateExamplePayload]
  cancel: []
}>()

const form = reactive({
  name: '',
  description: '',
})

function resetForm() {
  form.name = ''
  form.description = ''
}

watch(
  () => props.editingItem,
  (item) => {
    if (item) {
      form.name = item.name
      form.description = item.description
    } else {
      resetForm()
    }
  },
  { immediate: true },
)

function handleSubmit() {
  if (!form.name.trim()) return
  emit('submit', {
    name: form.name.trim(),
    description: form.description.trim(),
  })
  if (!props.editingItem) {
    resetForm()
  }
}
</script>

<template>
  <form
    class="rounded-xl border border-slate-200 bg-white p-5 shadow-xs"
    @submit.prevent="handleSubmit"
  >
    <div class="mb-4 flex items-center justify-between">
      <h2 class="text-sm font-semibold text-slate-900">
        {{ editingItem ? 'Edit Example' : 'Create Example' }}
      </h2>
      <BaseButton v-if="editingItem" variant="ghost" size="sm" @click="emit('cancel')">
        Cancel edit
      </BaseButton>
    </div>

    <div class="space-y-4">
      <div>
        <label class="mb-1 block text-xs font-medium text-slate-700" for="example-name">
          Name
        </label>
        <input
          id="example-name"
          v-model="form.name"
          type="text"
          required
          maxlength="200"
          placeholder="Enter example name..."
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 focus:outline-none"
        />
      </div>

      <div>
        <label class="mb-1 block text-xs font-medium text-slate-700" for="example-desc">
          Description
        </label>
        <textarea
          id="example-desc"
          v-model="form.description"
          rows="2"
          placeholder="Optional description..."
          class="w-full rounded-lg border border-slate-300 px-3 py-2 text-sm text-slate-900 placeholder:text-slate-400 focus:border-indigo-500 focus:ring-2 focus:ring-indigo-500/20 focus:outline-none"
        />
      </div>
    </div>

    <div class="mt-4 flex justify-end">
      <BaseButton type="submit" :disabled="submitting || !form.name.trim()">
        {{ submitting ? 'Saving...' : editingItem ? 'Save Changes' : 'Create Example' }}
      </BaseButton>
    </div>
  </form>
</template>
