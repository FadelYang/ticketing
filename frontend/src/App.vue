<script setup lang="ts">
import { computed } from 'vue'
import { useRoute } from 'vue-router'
import AppNavbar from '@/components/layout/AppNavbar.vue'

const route = useRoute()
const hideNavbar = computed(() => Boolean(route.meta.hideNavbar))
</script>

<template>
  <div class="min-h-screen bg-surface text-on-surface">
    <template v-if="hideNavbar">
      <RouterView />
    </template>
    <template v-else>
      <AppNavbar />
      <main class="mx-auto max-w-4xl px-4 py-8 sm:px-6 lg:px-8">
        <RouterView />
      </main>
    </template>
  </div>
</template>
