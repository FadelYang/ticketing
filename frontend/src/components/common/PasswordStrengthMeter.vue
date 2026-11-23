<script setup lang="ts">
import { computed } from 'vue'
import BaseIcon from '@/components/common/BaseIcon.vue'

interface Props {
  password: string
  helperText?: string
}

const props = withDefaults(defineProps<Props>(), {
  helperText: 'Must contain at least 8 characters, 1 digit, and 1 symbol',
})

const strength = computed(() => {
  const val = props.password
  if (!val) {
    return { score: 0, label: '8+ CHARACTERS', textClass: 'text-outline', barClass: '' }
  }

  let score = 0
  if (val.length >= 8) score++
  if (/[0-9]/.test(val)) score++
  if (/[^A-Za-z0-9]/.test(val)) score++
  if (val.length >= 12 && /[A-Z]/.test(val)) score++

  if (score <= 1) {
    return {
      score: 1,
      label: 'WEAK',
      textClass: 'text-error font-medium',
      barClass: 'bg-error',
    }
  }
  if (score === 2) {
    return {
      score: 2,
      label: 'FAIR',
      textClass: 'text-secondary font-medium',
      barClass: 'bg-secondary',
    }
  }
  if (score === 3) {
    return {
      score: 3,
      label: 'GOOD',
      textClass: 'text-primary font-medium',
      barClass: 'bg-primary',
    }
  }
  return {
    score: 4,
    label: 'STRONG',
    textClass: 'text-primary-container font-semibold',
    barClass: 'bg-primary-container',
  }
})

defineExpose({ strength })
</script>

<template>
  <div class="flex flex-col gap-1.5">
    <div class="w-full grid grid-cols-4 gap-1.5 pt-1">
      <div
        v-for="bar in 4"
        :key="bar"
        class="h-1 rounded-full transition-colors duration-200"
        :class="bar <= strength.score ? strength.barClass : 'bg-surface-container-highest'"
      />
    </div>
    <p class="font-body-sm text-body-sm text-on-surface-variant flex items-center gap-1 pt-0.5">
      <BaseIcon name="info" class="text-body-sm text-outline" />
      <span>{{ helperText }}</span>
    </p>
  </div>
</template>
