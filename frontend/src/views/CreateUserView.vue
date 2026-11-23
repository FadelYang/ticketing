<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ApiError } from '@/services/api'
import { userService } from '@/services/userService'
import { useTicketStore } from '@/stores/ticketStore'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import BaseAlert from '@/components/common/BaseAlert.vue'
import BaseInput from '@/components/common/BaseInput.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'
import PasswordStrengthMeter from '@/components/common/PasswordStrengthMeter.vue'
import type { UserRole } from '@/types/auth'

const router = useRouter()
const ticketStore = useTicketStore()
const { statusCounts } = storeToRefs(ticketStore)

const name = ref('')
const email = ref('')
const phonenumber = ref('')
const role = ref<UserRole>('agent')
const password = ref('')
const confirmPassword = ref('')
const loading = ref(false)
const error = ref<string | null>(null)
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
  void ticketStore.fetchStatusCounts()
})

async function handleSubmit() {
  mismatchError.value = ''
  error.value = null

  if (password.value !== confirmPassword.value) {
    mismatchError.value = 'Passwords do not match. Please verify your credentials.'
    return
  }

  loading.value = true
  try {
    await userService.create({
      name: name.value,
      email: email.value,
      phonenumber: phonenumber.value,
      password: password.value,
      role: role.value,
    })
    await router.push('/team')
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Failed to create user.'
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <DashboardLayout :total-count="statusCounts.all">
    <div class="flex flex-col w-full pb-space-2xl">
      <!-- Top Navigation Context -->
      <div
        class="flex flex-col sm:flex-row sm:items-center justify-between gap-space-md mb-space-xl pt-space-lg"
      >
        <div class="flex items-center gap-space-sm">
          <RouterLink
            to="/team"
            class="inline-flex items-center gap-space-xs font-label-md text-label-md text-on-surface-variant hover:text-primary transition-colors py-1 px-2 rounded-lg bg-surface-container-low hover:bg-surface-container"
          >
            <BaseIcon name="arrow_back" class="text-[18px]" />
            <span>Back to Team</span>
          </RouterLink>
          <span class="text-outline-variant text-[12px]">/</span>
          <span class="font-code-sm text-code-sm text-outline uppercase tracking-wider">
            NEW USER
          </span>
        </div>
      </div>

      <!-- Primary Page Header -->
      <div class="flex flex-col gap-1 mb-space-xl">
        <h1 class="font-headline-xl text-headline-xl text-on-surface font-bold tracking-tight">
          Create User
        </h1>
      </div>

      <BaseAlert v-if="error" variant="error" :message="error" class="mb-space-lg" />

      <!-- Master Form Container (Ticket-style Layout with Register Form Fields) -->
      <form class="flex flex-col gap-space-xl" @submit.prevent="handleSubmit">
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-space-xl items-start">
          <!-- Left Column (Account Credentials) - 8 cols -->
          <div class="lg:col-span-8 flex flex-col gap-space-lg">
            <div class="bg-surface-container-lowest rounded-xl p-space-xl shadow-sm flex flex-col gap-space-md">
              <BaseInput
                id="userFullName"
                v-model="name"
                label="Full Name"
                type="text"
                placeholder="Jane Doe"
                autocomplete="name"
                leading-icon="person"
                surface="low"
                required
              >
                <template #header-right>
                  <span class="font-code-sm text-code-sm text-outline">REQUIRED</span>
                </template>
              </BaseInput>

              <BaseInput
                id="userWorkEmail"
                v-model="email"
                label="Work Email"
                type="email"
                placeholder="name@company.com"
                autocomplete="email"
                leading-icon="mail"
                surface="low"
                required
              >
                <template #header-right>
                  <span class="font-code-sm text-code-sm text-outline">REQUIRED</span>
                </template>
              </BaseInput>

              <BaseInput
                id="userPhoneNumber"
                v-model="phonenumber"
                label="Phone Number"
                type="tel"
                placeholder="085156273182"
                autocomplete="tel"
                leading-icon="call"
                surface="low"
                numeric-only
                required
              >
                <template #header-right>
                  <span class="font-code-sm text-code-sm text-outline">REQUIRED</span>
                </template>
              </BaseInput>

              <BaseInput
                id="createUserPassword"
                v-model="password"
                label="Password"
                type="password"
                placeholder="••••••••••••"
                autocomplete="new-password"
                leading-icon="lock"
                password-toggle
                monospace
                surface="low"
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
                id="createConfirmPassword"
                v-model="confirmPassword"
                label="Confirm Password"
                type="password"
                placeholder="••••••••••••"
                autocomplete="new-password"
                leading-icon="lock_reset"
                password-toggle
                monospace
                surface="low"
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
            </div>
          </div>

          <!-- Right Column (Role Assignment) - 4 cols -->
          <div class="lg:col-span-4 flex flex-col gap-space-lg">
            <div class="bg-surface-container-lowest rounded-xl p-space-xl shadow-sm flex flex-col gap-space-md">
              <div class="flex flex-col gap-space-sm">
                <label
                  for="user-role"
                  class="font-label-md text-label-md text-on-surface font-semibold flex items-center gap-1"
                >
                  Role <span class="text-error">*</span>
                </label>
                <div class="relative">
                  <select
                    id="user-role"
                    v-model="role"
                    required
                    class="w-full h-10 pl-3 pr-8 rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md appearance-none focus:outline-none focus:bg-surface-container-lowest focus:shadow-[0_0_0_2px_rgba(79,70,229,0.25)] transition-all cursor-pointer"
                  >
                    <option value="requester">Requester</option>
                    <option value="agent">Agent</option>
                    <option value="admin">Admin</option>
                  </select>
                  <BaseIcon
                    name="unfold_more"
                    class="absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none text-outline text-[18px]"
                  />
                </div>
                <p class="font-body-sm text-body-sm text-outline">
                  Admins can manage users and tickets; Agents handle assigned tickets; Requesters
                  submit tickets.
                </p>
              </div>
            </div>
          </div>
        </div>

        <!-- Sticky Bottom Action Footer -->
        <div
          class="sticky bottom-0 z-30 -mx-gutter-desktop px-gutter-desktop py-space-md bg-surface-container-lowest/90 backdrop-blur-md shadow-md rounded-t-xl flex flex-col sm:flex-row items-center justify-end gap-space-md mt-space-lg"
        >
          <div class="flex items-center gap-space-md w-full sm:w-auto justify-end">
            <RouterLink
              to="/team"
              class="px-4 py-2 rounded-lg font-label-md text-label-md text-on-surface-variant hover:text-on-surface hover:bg-surface-container transition-colors"
            >
              Cancel
            </RouterLink>
            <button
              type="submit"
              :disabled="loading"
              class="inline-flex items-center gap-2 px-5 py-2 rounded-lg bg-primary-container hover:bg-primary text-on-primary font-label-md text-label-md font-semibold shadow-sm transition-all cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed"
            >
              <BaseIcon v-if="loading" name="progress_activity" class="text-[18px] animate-spin" />
              <span>{{ loading ? 'Creating...' : 'Create User' }}</span>
              <span v-if="!loading" class="font-bold text-[15px]">→</span>
            </button>
          </div>
        </div>
      </form>
    </div>
  </DashboardLayout>
</template>
