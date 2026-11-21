<script setup lang="ts">
import AppNavbar from '@/components/layout/AppNavbar.vue'
</script>

<template>
  <div class="min-h-screen bg-slate-50 text-slate-900">
    <AppNavbar />
    <main class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
      <RouterView />
    </main>
  </div>
</template>
