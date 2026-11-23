<script setup lang="ts">
import { computed, ref } from 'vue'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import AuthCard from '@/components/auth/AuthCard.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'
import PasswordStrengthMeter from '@/components/common/PasswordStrengthMeter.vue'

const email = ref('')
const password = ref('')
const confirmPassword = ref('')
const status = ref<'idle' | 'submitting' | 'created'>('idle')
const mismatchError = ref('')

const strengthRef = ref<InstanceType<typeof PasswordStrengthMeter> | null>(null)

const strengthLabel = computed(() => strengthRef.value?.strength.label ?? '8+ CHARACTERS')
const strengthTextClass = computed(() => strengthRef.value?.strength.textClass ?? 'text-outline')

const matchStatus = computed(() => {
  if (!confirmPassword.value) return null
  if (password.value === confirmPassword.value) {
    return { text: 'MATCHED', className: 'text-primary font-medium' }
  }
  return { text: 'MISMATCH', className: 'text-error font-medium' }
})

function handleRegistration() {
  mismatchError.value = ''
  if (password.value !== confirmPassword.value) {
    mismatchError.value = 'Passwords do not match. Please verify your credentials.'
    return
  }

  status.value = 'submitting'
  // UI-only preview state until auth API is wired
  setTimeout(() => {
    status.value = 'created'
  }, 1200)
}
</script>

<template>
  <AuthLayout max-width="md">
    <AuthCard
      brand-name="Simple Ticketing"
      title="Create your account"
      subtitle="Get started tracking and resolving tickets in seconds"
      header-align="center"
      show-accent-bar
      show-system-bar
      footer-variant="band"
    >
      <template #brand-icon>
        <div
          class="flex items-center justify-center w-12 h-12 rounded-lg bg-surface-container text-primary mb-space-md shadow-sm"
        >
          <BaseIcon name="confirmation_number" class="text-headline-xl" />
        </div>
      </template>

      <form class="flex flex-col space-y-space-md" @submit.prevent="handleRegistration">
        <BaseInput
          id="workEmail"
          v-model="email"
          label="Work Email"
          type="email"
          placeholder="name@company.com"
          autocomplete="email"
          leading-icon="mail"
          surface="lowest"
          required
        >
          <template #header-right>
            <span class="font-code-sm text-code-sm text-outline">REQUIRED</span>
          </template>
        </BaseInput>

        <BaseInput
          id="userPassword"
          v-model="password"
          label="Password"
          type="password"
          placeholder="••••••••••••"
          autocomplete="new-password"
          leading-icon="lock"
          password-toggle
          monospace
          surface="lowest"
          required
        >
          <template #header-right>
            <span class="font-code-sm text-code-sm uppercase" :class="strengthTextClass">
              {{ strengthLabel }}
            </span>
          </template>
          <template #helper>
            <PasswordStrengthMeter ref="strengthRef" :password="password" />
          </template>
        </BaseInput>

        <BaseInput
          id="confirmPassword"
          v-model="confirmPassword"
          label="Confirm Password"
          type="password"
          placeholder="••••••••••••"
          autocomplete="new-password"
          leading-icon="lock_reset"
          password-toggle
          monospace
          surface="lowest"
          :error="mismatchError"
          required
        >
          <template #header-right>
            <span
              v-if="matchStatus"
              class="font-code-sm text-code-sm"
              :class="matchStatus.className"
            >
              {{ matchStatus.text }}
            </span>
          </template>
        </BaseInput>

        <div class="pt-space-xs">
          <p class="font-body-sm text-body-sm text-on-surface-variant leading-relaxed">
            By creating an account, you agree to our
            <a class="text-primary font-medium hover:underline focus:outline-none" href="#terms">
              Terms of Service
            </a>
            and acknowledge our
            <a class="text-primary font-medium hover:underline focus:outline-none" href="#privacy">
              Privacy Policy
            </a>
            .
          </p>
        </div>

        <div class="pt-space-xs">
          <BaseButton
            type="submit"
            block
            :variant="status === 'created' ? 'success' : 'primary'"
            :loading="status === 'submitting'"
            :disabled="status === 'created'"
          >
            <template v-if="status === 'submitting'">
              <span>Provisioning Account...</span>
            </template>
            <template v-else-if="status === 'created'">
              <BaseIcon name="check_circle" class="text-headline-sm" />
              <span>Account Created</span>
            </template>
            <template v-else>
              <span>Create Account</span>
              <BaseIcon name="arrow_forward" class="text-headline-sm" />
            </template>
          </BaseButton>
        </div>
      </form>

      <template #footer>
        <span class="font-body-md text-body-md text-on-surface-variant">
          Already have an account?
        </span>
        <RouterLink
          to="/login"
          class="ml-1.5 font-label-md text-label-md text-primary font-semibold hover:underline flex items-center space-x-0.5 focus:outline-none"
        >
          <span>Log in</span>
          <BaseIcon name="login" class="text-body-md" />
        </RouterLink>
      </template>
    </AuthCard>
  </AuthLayout>
</template>
