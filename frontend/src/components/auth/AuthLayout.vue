<script setup lang="ts">
interface Props {
  maxWidth?: 'sm' | 'md'
}

withDefaults(defineProps<Props>(), {
  maxWidth: 'md',
})
</script>

<template>
  <main
    class="min-h-screen w-full flex flex-col items-center justify-center p-gutter-desktop relative overflow-hidden"
  >
    <!-- Ambient backdrop lighting -->
    <div class="fixed inset-0 pointer-events-none overflow-hidden flex items-center justify-center opacity-60">
      <div
        class="w-[600px] h-[600px] rounded-full bg-surface-container-high/40 blur-3xl -top-40 -right-20 absolute"
      />
      <div
        class="w-[500px] h-[500px] rounded-full bg-primary-fixed/20 blur-3xl -bottom-20 -left-20 absolute"
      />
    </div>

    <div
      class="relative w-full mx-auto z-10"
      :class="maxWidth === 'sm' ? 'max-w-md' : 'max-w-[460px]'"
    >
      <slot />
    </div>
  </main>
</template>
