<script setup lang="ts">
interface Props {
  name: string
}

defineProps<Props>()
</script>

<template>
  <span class="material-symbols-outlined leading-none select-none" aria-hidden="true">{{ name }}</span>
</template>
