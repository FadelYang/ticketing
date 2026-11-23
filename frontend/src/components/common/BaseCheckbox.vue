<script setup lang="ts">
interface Props {
  modelValue: boolean
  id: string
  label: string
}

const props = defineProps<Props>()

const emit = defineEmits<{
  'update:modelValue': [value: boolean]
}>()

function onChange(event: Event) {
  emit('update:modelValue', (event.target as HTMLInputElement).checked)
}
</script>

<template>
  <label :for="id" class="flex items-center gap-space-sm cursor-pointer select-none group">
    <input
      :id="id"
      type="checkbox"
      :checked="props.modelValue"
      class="sr-only peer"
      @change="onChange"
    />
    <div
      class="w-4 h-4 rounded bg-surface-container-high peer-checked:bg-primary-container flex items-center justify-center transition-all group-hover:bg-surface-container-highest peer-focus-visible:shadow-[0_0_0_2px] peer-focus-visible:shadow-primary-container"
    >
      <svg
        class="w-3 h-3 text-on-primary transition-opacity"
        :class="props.modelValue ? 'opacity-100' : 'opacity-0'"
        fill="none"
        stroke="currentColor"
        stroke-linecap="round"
        stroke-linejoin="round"
        stroke-width="3"
        viewBox="0 0 24 24"
      >
        <polyline points="20 6 9 17 4 12" />
      </svg>
    </div>
    <span class="font-body-sm text-body-sm text-on-surface-variant group-hover:text-on-surface transition-colors">
      {{ label }}
    </span>
  </label>
</template>
