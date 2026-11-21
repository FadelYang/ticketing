<script setup lang="ts">
interface Props {
  title: string
  description: string
}

defineProps<Props>()
</script>

<template>
  <div
    class="flex flex-col items-center justify-center rounded-xl border border-dashed border-slate-300 bg-white px-6 py-12 text-center"
  >
    <p class="text-base font-semibold text-slate-800">{{ title }}</p>
    <p class="mt-1 max-w-sm text-sm text-slate-500">{{ description }}</p>
  </div>
</template>
