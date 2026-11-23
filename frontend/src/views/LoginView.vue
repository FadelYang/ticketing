<script setup lang="ts">
import { ref } from 'vue'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import AuthCard from '@/components/auth/AuthCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseCheckbox from '@/components/common/BaseCheckbox.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'

const email = ref('')
const password = ref('')
const rememberMe = ref(false)
const submitting = ref(false)

function handleLogin() {
  submitting.value = true
  // UI-only preview state until auth API is wired
  setTimeout(() => {
    submitting.value = false
  }, 800)
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

        <BaseButton type="submit" block :loading="submitting" class="mt-space-sm">
          <span>{{ submitting ? 'Signing In...' : 'Sign In' }}</span>
          <BaseIcon v-if="!submitting" name="arrow_forward" class="text-base" />
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
