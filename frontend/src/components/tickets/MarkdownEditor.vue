<script setup lang="ts">
interface Props {
  modelValue: string
  id: string
  label?: string
  placeholder?: string
  rows?: number
  required?: boolean
}

withDefaults(defineProps<Props>(), {
  label: 'Description & Steps',
  placeholder: 'Describe the issue, steps to reproduce, and any relevant logs...',
  rows: 8,
  required: false,
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()
</script>

<template>
  <div class="flex flex-col gap-space-sm">
    <div class="flex items-center justify-between">
      <label
        :for="id"
        class="font-label-md text-label-md text-on-surface font-semibold flex items-center gap-1"
      >
        <span>{{ label }}</span>
        <span v-if="required" class="text-error">*</span>
      </label>
    </div>

    <textarea
      :id="id"
      :value="modelValue"
      :rows="rows"
      :placeholder="placeholder"
      :required="required"
      class="w-full p-space-md rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md placeholder:text-outline focus:outline-none focus:bg-surface-container-lowest focus:shadow-[0_0_0_2px_rgba(79,70,229,0.25)] transition-all resize-y"
      @input="emit('update:modelValue', ($event.target as HTMLTextAreaElement).value)"
    />
  </div>
</template>
