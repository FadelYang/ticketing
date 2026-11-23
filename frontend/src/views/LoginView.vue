<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/authStore'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import AuthCard from '@/components/auth/AuthCard.vue'
import BaseAlert from '@/components/common/BaseAlert.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseCheckbox from '@/components/common/BaseCheckbox.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'

const router = useRouter()
const route = useRoute()
const authStore = useAuthStore()
const { loading, error } = storeToRefs(authStore)

const email = ref('')
const password = ref('')
const rememberMe = ref(false)

onMounted(() => {
  authStore.clearError()
})

async function handleLogin() {
  const ok = await authStore.login(
    {
      email: email.value,
      password: password.value,
    },
    rememberMe.value,
  )

  if (ok) {
    const redirect = typeof route.query.redirect === 'string' ? route.query.redirect : '/tickets'
    await router.push(redirect)
  }
}
</script>

<template>
  <AuthLayout max-width="sm">
    <AuthCard
      brand-name="Simple Ticketing"
      title="Welcome back"
      subtitle="Log in to manage and track support tickets"
      header-align="left"
      footer-variant="plain"
    >
      <form class="flex flex-col gap-space-lg" @submit.prevent="handleLogin">
        <BaseAlert v-if="error" variant="error" :message="error" />

        <BaseInput
          id="work-email"
          v-model="email"
          label="Work Email"
          type="email"
          placeholder="name@company.com"
          autocomplete="email"
          trailing-icon="mail"
          required
        />

        <BaseInput
          id="password"
          v-model="password"
          label="Password"
          type="password"
          placeholder="••••••••••••"
          autocomplete="current-password"
          password-toggle
          required
        >
          <template #header-right>
            <a
              class="font-label-md text-label-md text-primary font-medium hover:text-primary-container transition-colors"
              href="#forgot-password"
            >
              Forgot password?
            </a>
          </template>
        </BaseInput>

        <div class="flex items-center justify-between pt-space-xs">
          <BaseCheckbox id="remember-me" v-model="rememberMe" label="Remember me for 30 days" />
        </div>

        <BaseButton type="submit" block :loading="loading" class="mt-space-sm">
          <span>{{ loading ? 'Signing In...' : 'Sign In' }}</span>
          <BaseIcon v-if="!loading" name="arrow_forward" class="text-base" />
        </BaseButton>

        <div class="relative flex items-center justify-center my-space-xs">
          <div class="w-full h-px bg-surface-container-high" />
        </div>
      </form>

      <template #footer>
        <p class="font-body-sm text-body-sm text-on-surface-variant">
          Don't have an account?
          <RouterLink
            to="/register"
            class="font-label-sm text-label-sm text-primary font-semibold hover:text-primary-container transition-colors ml-1"
          >
            Sign up
          </RouterLink>
        </p>
      </template>
    </AuthCard>
  </AuthLayout>
</template>
