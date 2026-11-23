<script setup lang="ts">
import { computed, ref } from 'vue'
import BaseIcon from '@/components/common/BaseIcon.vue'

interface Props {
  modelValue: string
  id: string
  label?: string
  type?: 'text' | 'email' | 'password' | 'search' | 'url' | 'tel'
  placeholder?: string
  autocomplete?: string
  required?: boolean
  disabled?: boolean
  leadingIcon?: string
  trailingIcon?: string
  passwordToggle?: boolean
  monospace?: boolean
  surface?: 'low' | 'lowest'
  numericOnly?: boolean
  error?: string
}

const props = withDefaults(defineProps<Props>(), {
  label: '',
  type: 'text',
  placeholder: '',
  autocomplete: undefined,
  required: false,
  disabled: false,
  leadingIcon: undefined,
  trailingIcon: undefined,
  passwordToggle: false,
  monospace: false,
  surface: 'low',
  numericOnly: false,
  error: '',
})

const emit = defineEmits<{
  'update:modelValue': [value: string]
}>()

const showPassword = ref(false)

const resolvedType = computed(() => {
  if (props.type === 'password' && props.passwordToggle) {
    return showPassword.value ? 'text' : 'password'
  }
  return props.type
})

const hasTrailingControl = computed(
  () => Boolean(props.trailingIcon) || (props.type === 'password' && props.passwordToggle),
)

const inputClasses = computed(() => {
  const base =
    'w-full h-10 rounded-lg text-on-surface placeholder:text-outline/70 font-body-md text-body-md transition-all outline-none disabled:cursor-not-allowed disabled:opacity-60'

  const paddingLeft = props.leadingIcon ? 'pl-11' : 'px-space-md'
  const paddingRight = hasTrailingControl.value ? 'pr-10' : 'pr-3'

  const surfaceStyle =
    props.surface === 'lowest'
      ? 'bg-surface-container-lowest shadow-sm focus:ring-2 focus:ring-primary focus:bg-surface-container-low'
      : 'bg-surface-container-low focus:bg-surface-container-lowest focus:shadow-[0_0_0_2px] focus:shadow-primary-container'

  const errorStyle = props.error ? 'ring-2 ring-error' : ''
  const fontStyle = props.monospace ? 'font-code-sm' : ''

  return [base, paddingLeft, paddingRight, surfaceStyle, errorStyle, fontStyle]
    .filter(Boolean)
    .join(' ')
})

function onInput(event: Event) {
  const target = event.target as HTMLInputElement
  let value = target.value
  if (props.numericOnly) {
    value = value.replace(/\D/g, '')
    if (target.value !== value) {
      target.value = value
    }
  }
  emit('update:modelValue', value)
}
</script>

<template>
  <div class="flex flex-col gap-space-sm">
    <div v-if="label || $slots['header-right']" class="flex items-center justify-between">
      <label v-if="label" :for="id" class="font-label-md text-label-md text-on-surface font-semibold">
        {{ label }}
      </label>
      <slot name="header-right" />
    </div>

    <div class="relative flex items-center">
      <BaseIcon
        v-if="leadingIcon"
        :name="leadingIcon"
        class="absolute left-3 text-outline text-lg pointer-events-none"
      />

      <input
        :id="id"
        :type="resolvedType"
        :value="modelValue"
        :placeholder="placeholder"
        :autocomplete="autocomplete"
        :required="required"
        :disabled="disabled"
        :inputmode="numericOnly ? 'numeric' : undefined"
        :class="inputClasses"
        @input="onInput"
      />

      <button
        v-if="type === 'password' && passwordToggle"
        type="button"
        aria-label="Toggle password visibility"
        class="absolute right-2.5 p-1 flex items-center justify-center text-outline hover:text-on-surface transition-colors cursor-pointer rounded focus:outline-none"
        @click="showPassword = !showPassword"
      >
        <BaseIcon :name="showPassword ? 'visibility_off' : 'visibility'" class="text-lg" />
      </button>

      <BaseIcon
        v-else-if="trailingIcon"
        :name="trailingIcon"
        class="absolute right-3 text-outline text-lg pointer-events-none"
      />
    </div>

    <p v-if="error" class="font-body-sm text-body-sm text-error">
      {{ error }}
    </p>
    <slot name="helper" />
  </div>
</template>
