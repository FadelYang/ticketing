<script setup lang="ts">
import { computed } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/authStore'
import BaseIcon from '@/components/common/BaseIcon.vue'

interface Props {
  totalCount?: number
  searchQuery?: string
}

withDefaults(defineProps<Props>(), {
  totalCount: 0,
  searchQuery: '',
})

const emit = defineEmits<{
  'update:searchQuery': [value: string]
}>()

const router = useRouter()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)

const roleLabel = computed(() => {
  const role = user.value?.role
  if (role === 'admin') return 'Administrator'
  if (role === 'agent') return 'Support Agent'
  return 'Requester'
})

const sidebarLabel = computed(() =>
  user.value?.role === 'requester' ? 'My Tickets' : 'All Tickets',
)

const userInitials = computed(() => {
  const name = user.value?.name?.trim()
  if (!name) return 'U'
  const parts = name.split(/\s+/)
  return parts.length >= 2
    ? `${parts[0][0]}${parts[1][0]}`.toUpperCase()
    : name.slice(0, 2).toUpperCase()
})

async function handleLogout() {
  await authStore.logout()
  await router.push('/login')
}
</script>

<template>
  <div class="bg-surface font-body-md text-on-surface antialiased min-h-screen">
    <!-- Fixed Top Header -->
    <header
      class="fixed top-0 left-0 right-0 h-16 bg-surface-container-lowest border-b border-outline-variant/30 z-50 flex items-center justify-between px-gutter-desktop"
    >
      <div class="flex items-center gap-space-xl">
        <RouterLink to="/tickets" class="flex items-center gap-space-sm cursor-pointer">
          <div class="w-8 h-8 rounded-lg bg-primary flex items-center justify-center">
            <BaseIcon name="confirmation_number" class="text-on-primary text-[20px]" />
          </div>
          <span class="font-headline-md text-headline-md text-on-surface tracking-tight font-bold">
            DeskFlow
          </span>
        </RouterLink>

        <div class="h-4 w-px bg-outline-variant/40 hidden md:block" />

        <nav class="hidden md:flex items-center gap-space-lg h-16">
          <RouterLink
            to="/tickets"
            class="h-full flex items-center px-space-xs transition-colors text-primary border-b-2 border-primary font-semibold"
          >
            Tickets
          </RouterLink>
          <a
            v-if="user?.role === 'admin' || user?.role === 'agent'"
            class="h-full flex items-center text-on-surface-variant hover:text-on-surface font-body-md text-body-md px-space-xs transition-colors"
            href="#team"
            @click.prevent
          >
            Team
          </a>
          <a
            class="h-full flex items-center text-on-surface-variant hover:text-on-surface font-body-md text-body-md px-space-xs transition-colors"
            href="#settings"
            @click.prevent
          >
            Settings
          </a>
        </nav>
      </div>

      <div class="flex items-center gap-space-md">
        <div class="relative hidden sm:flex items-center">
          <BaseIcon
            name="search"
            class="absolute left-space-sm text-outline text-[18px] pointer-events-none top-1/2 -translate-y-1/2"
          />
          <input
            :value="searchQuery"
            type="text"
            placeholder="Search tickets, queues, or customers..."
            class="w-72 pl-8 pr-12 h-8 bg-surface-container-low border border-outline-variant/50 rounded-lg text-on-surface placeholder:text-outline font-body-sm text-body-sm focus:outline-none focus:border-primary focus:bg-surface-container-lowest transition-all"
            @input="emit('update:searchQuery', ($event.target as HTMLInputElement).value)"
          />
          <kbd
            class="absolute right-space-sm top-1/2 -translate-y-1/2 px-1.5 py-0.5 rounded bg-surface-container-high border border-outline-variant/40 text-on-surface-variant font-code-sm text-[10px] leading-tight flex items-center"
          >
            ⌘K
          </kbd>
        </div>

        <button
          type="button"
          class="relative w-8 h-8 rounded-lg text-on-surface-variant hover:bg-surface-container-high hover:text-on-surface transition-colors flex items-center justify-center cursor-pointer"
          title="Notifications"
        >
          <BaseIcon name="notifications" class="text-[20px]" />
          <span class="absolute top-1.5 right-1.5 w-2 h-2 bg-primary rounded-full" />
        </button>

        <div class="h-5 w-px bg-outline-variant/40" />

        <div class="flex items-center gap-space-sm">
          <div
            class="w-8 h-8 rounded-full bg-primary flex items-center justify-center text-on-primary text-xs font-semibold"
            :title="user ? `${user.name} (${roleLabel})` : 'User'"
          >
            {{ userInitials }}
          </div>
          <div class="hidden xl:flex flex-col leading-tight">
            <span class="text-label-md font-semibold text-on-surface">{{ user?.name || 'User' }}</span>
            <span class="text-[11px] text-outline capitalize">{{ roleLabel }}</span>
          </div>
          <button
            type="button"
            title="Log out"
            class="ml-1 p-1.5 rounded-lg text-on-surface-variant hover:bg-surface-container-high hover:text-on-surface transition-colors flex items-center justify-center cursor-pointer"
            @click="handleLogout"
          >
            <BaseIcon name="logout" class="text-[18px]" />
          </button>
        </div>
      </div>
    </header>

    <!-- Fixed Left Sidebar -->
    <aside
      class="fixed left-0 top-16 bottom-0 w-64 bg-surface-container-lowest border-r border-outline-variant/30 z-40 hidden lg:flex flex-col justify-between p-space-lg"
    >
      <div class="flex flex-col gap-space-lg">
        <nav class="flex flex-col gap-1">
          <RouterLink
            to="/tickets"
            class="flex items-center justify-between px-space-md py-space-sm rounded-lg transition-colors bg-primary-container text-on-primary font-semibold shadow-sm"
          >
            <div class="flex items-center gap-space-sm">
              <BaseIcon name="inbox" class="text-[18px]" />
              <span>{{ sidebarLabel }}</span>
            </div>
            <span
              class="font-code-sm text-code-sm bg-surface-container px-1.5 py-0.5 rounded-full text-on-surface-variant"
            >
              {{ totalCount }}
            </span>
          </RouterLink>
        </nav>
      </div>
    </aside>

    <!-- Main Workspace -->
    <div class="lg:pl-64">
      <main class="w-full pt-20 bg-surface min-h-screen px-gutter-desktop pb-space-2xl">
        <slot />
      </main>
    </div>
  </div>
</template>
