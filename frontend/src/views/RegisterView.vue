<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { useAuthStore } from '@/stores/authStore'
import AuthLayout from '@/components/auth/AuthLayout.vue'
import AuthCard from '@/components/auth/AuthCard.vue'
import BaseAlert from '@/components/common/BaseAlert.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseButton from '@/components/common/BaseButton.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'
import PasswordStrengthMeter from '@/components/common/PasswordStrengthMeter.vue'

const router = useRouter()
const authStore = useAuthStore()
const { loading, error } = storeToRefs(authStore)

const name = ref('')
const email = ref('')
const phonenumber = ref('')
const password = ref('')
const confirmPassword = ref('')
const created = ref(false)
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

onMounted(() => {
  authStore.clearError()
})

async function handleRegistration() {
  mismatchError.value = ''
  authStore.clearError()

  if (password.value !== confirmPassword.value) {
    mismatchError.value = 'Passwords do not match. Please verify your credentials.'
    return
  }

  const ok = await authStore.register({
    name: name.value,
    email: email.value,
    phonenumber: phonenumber.value,
    password: password.value,
  })

  if (ok) {
    created.value = true
    setTimeout(() => {
      router.push('/')
    }, 600)
  }
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
        <BaseAlert v-if="error" variant="error" :message="error" />

        <BaseInput
          id="fullName"
          v-model="name"
          label="Full Name"
          type="text"
          placeholder="Jane Doe"
          autocomplete="name"
          leading-icon="person"
          surface="lowest"
          required
        >
          <template #header-right>
            <span class="font-code-sm text-code-sm text-outline">REQUIRED</span>
          </template>
        </BaseInput>

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
          id="phoneNumber"
          v-model="phonenumber"
          label="Phone Number"
          type="tel"
          placeholder="085156273182"
          autocomplete="tel"
          leading-icon="call"
          surface="lowest"
          numeric-only
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
            :variant="created ? 'success' : 'primary'"
            :loading="loading"
            :disabled="created"
          >
            <template v-if="loading">
              <span>Provisioning Account...</span>
            </template>
            <template v-else-if="created">
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
          class="ml-1.5 font-label-md text-label-md text-primary font-semibold hover:underline flex items-center gap-1 focus:outline-none"
        >
          <span>Log in</span>
          <BaseIcon name="login" class="text-body-md" />
        </RouterLink>
      </template>
    </AuthCard>
  </AuthLayout>
</template>
