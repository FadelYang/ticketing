<script setup lang="ts">
interface Props {
  brandName?: string
  title: string
  subtitle: string
  headerAlign?: 'left' | 'center'
  showAccentBar?: boolean
  showSystemBar?: boolean
  clusterLabel?: string
  versionLabel?: string
  footerVariant?: 'plain' | 'band'
}

withDefaults(defineProps<Props>(), {
  brandName: 'Simple Ticketing',
  headerAlign: 'left',
  showAccentBar: false,
  showSystemBar: false,
  footerVariant: 'plain',
})
</script>

<template>
  <div class="w-full flex flex-col">
    <!-- Optional System Architecture Header -->
    <div
      v-if="showSystemBar"
      class="flex items-center justify-between px-space-xs mb-space-lg text-tertiary"
    >
      <div
        class="flex items-center space-x-space-xs font-code-sm text-code-sm uppercase tracking-wider text-outline"
      >
        <span>{{ clusterLabel }}</span>
      </div>
      <div class="font-code-sm text-code-sm text-outline">
        <span>{{ versionLabel }}</span>
      </div>
    </div>

    <!-- Main Card -->
    <div
      class="bg-surface-container-lowest shadow-xl rounded-xl p-space-2xl sm:p-10 flex flex-col w-full relative overflow-hidden"
    >
      <!-- Optional Top Gradient Accent Bar -->
      <div
        v-if="showAccentBar"
        class="absolute top-0 left-0 right-0 h-1 bg-gradient-to-r from-primary to-secondary-container"
      />

      <!-- Brand & Header (Left-aligned) -->
      <template v-if="headerAlign === 'left'">
        <div class="flex items-center gap-space-sm mb-space-2xl self-start">
          <slot name="brand-icon">
            <div
              class="w-8 h-8 rounded-lg bg-primary-container flex items-center justify-center text-on-primary shadow-sm"
            >
              <svg
                class="w-4 h-4"
                fill="none"
                stroke="currentColor"
                stroke-linecap="round"
                stroke-linejoin="round"
                stroke-width="2.2"
                viewBox="0 0 24 24"
              >
                <polyline points="16 18 22 12 16 6" />
                <polyline points="8 6 2 12 8 18" />
              </svg>
            </div>
          </slot>
          <span class="font-headline-sm text-headline-sm text-on-surface tracking-tight">
            {{ brandName }}
          </span>
        </div>

        <div class="flex flex-col gap-space-xs mb-space-xl">
          <h1 class="font-headline-lg text-headline-lg text-on-surface">{{ title }}</h1>
          <p class="font-body-md text-body-md text-on-surface-variant">{{ subtitle }}</p>
        </div>
      </template>

      <!-- Brand & Header (Centered) -->
      <template v-else>
        <div class="flex flex-col items-center text-center mb-space-xl">
          <slot name="brand-icon" />
          <div class="flex items-center space-x-1.5 mb-space-xs">
            <span class="font-headline-sm text-headline-sm text-on-surface tracking-tight uppercase">
              {{ brandName }}
            </span>
          </div>
          <h1 class="font-headline-lg text-headline-lg text-on-surface mb-space-xs tracking-tight">
            {{ title }}
          </h1>
          <p class="font-body-md text-body-md text-on-surface-variant max-w-xs">
            {{ subtitle }}
          </p>
        </div>
      </template>

      <!-- Form / Main Content -->
      <slot />

      <!-- Footer -->
      <div
        v-if="$slots.footer"
        :class="
          footerVariant === 'band'
            ? 'mt-space-xl pt-space-md flex items-center justify-center bg-surface-container-low -mx-space-2xl sm:-mx-10 -mb-space-2xl sm:-mb-10 p-space-md'
            : 'mt-space-2xl pt-space-md text-center'
        "
      >
        <slot name="footer" />
      </div>
    </div>
  </div>
</template>
