<script setup lang="ts">
import { onMounted, onUnmounted, ref } from 'vue'
import { useRouter } from 'vue-router'
import { storeToRefs } from 'pinia'
import { attachmentService } from '@/services/attachmentService'
import { useAuthStore } from '@/stores/authStore'
import { useTicketStore } from '@/stores/ticketStore'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import MarkdownEditor from '@/components/tickets/MarkdownEditor.vue'
import PrioritySelector from '@/components/tickets/PrioritySelector.vue'
import BaseAlert from '@/components/common/BaseAlert.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'
import type { TicketPriority } from '@/types/ticket'

const router = useRouter()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)

const ticketStore = useTicketStore()
const { total, loading, error } = storeToRefs(ticketStore)

const title = ref('')
const description = ref('')
const category = ref('')
const priority = ref<TicketPriority>('medium')
const deadlineLocal = ref('')

function formatLocalDateTimeInput(date: Date): string {
  const offsetMs = date.getTimezoneOffset() * 60000
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16)
}

const minDeadlineLocal = ref(formatLocalDateTimeInput(new Date()))

interface PendingImage {
  file: File
  previewUrl: string
}

const fileInputRef = ref<HTMLInputElement | null>(null)
const pendingImages = ref<PendingImage[]>([])
const previewImageUrl = ref<string | null>(null)

function handleFilesSelected(event: Event) {
  const input = event.target as HTMLInputElement
  if (!input.files) return

  const next = [...pendingImages.value]
  for (const file of Array.from(input.files)) {
    if (!file.type.startsWith('image/')) continue
    if (next.length >= 5) break
    next.push({
      file,
      previewUrl: URL.createObjectURL(file),
    })
  }
  pendingImages.value = next
  input.value = ''
}

function removePendingImage(index: number) {
  const target = pendingImages.value[index]
  if (target) {
    URL.revokeObjectURL(target.previewUrl)
  }
  pendingImages.value = pendingImages.value.filter((_, idx) => idx !== index)
}

onUnmounted(() => {
  for (const item of pendingImages.value) {
    URL.revokeObjectURL(item.previewUrl)
  }
})

const categories = [
  'Infrastructure',
  'API & Gateway',
  'Database',
  'Authentication',
  'Billing & Payments',
  'Frontend UI',
]

onMounted(() => {
  ticketStore.clearError()
  minDeadlineLocal.value = formatLocalDateTimeInput(new Date())
})

async function handleSubmit() {
  if (!user.value) return

  const parsedDeadline = new Date(deadlineLocal.value)
  if (Number.isNaN(parsedDeadline.getTime())) {
    error.value = 'Please select a valid deadline.'
    return
  }
  if (parsedDeadline.getTime() < Date.now()) {
    minDeadlineLocal.value = formatLocalDateTimeInput(new Date())
    error.value = 'Deadline cannot be in the past.'
    return
  }

  const created = await ticketStore.createTicket({
    title: title.value,
    description: description.value,
    category: category.value,
    priority: priority.value,
    requester_id: user.value.id,
    assignee_id: null,
    deadline_at: parsedDeadline.toISOString(),
  })

  if (created) {
    if (pendingImages.value.length > 0) {
      try {
        await attachmentService.upload(
          'ticket',
          created.id,
          pendingImages.value.map((item) => item.file),
        )
      } catch {
        // Ticket was already created; navigate to detail view even if attachment upload fails
      }
    }
    await router.push('/tickets')
  }
}
</script>

<template>
  <DashboardLayout :total-count="total">
    <div class="flex flex-col w-full pb-space-2xl">
      <!-- Top Navigation Context -->
      <div
        class="flex flex-col sm:flex-row sm:items-center justify-between gap-space-md mb-space-xl pt-space-lg"
      >
        <div class="flex items-center gap-space-sm">
          <RouterLink
            to="/tickets"
            class="inline-flex items-center gap-space-xs font-label-md text-label-md text-on-surface-variant hover:text-primary transition-colors py-1 px-2 rounded-lg bg-surface-container-low hover:bg-surface-container"
          >
            <BaseIcon name="arrow_back" class="text-[18px]" />
            <span>Back to Tickets</span>
          </RouterLink>
          <span class="text-outline-variant text-[12px]">/</span>
          <span class="font-code-sm text-code-sm text-outline uppercase tracking-wider">
            NEW ISSUE
          </span>
        </div>

        <span
          class="px-2.5 py-0.5 rounded-full bg-secondary-fixed text-on-secondary-fixed font-label-sm text-label-sm uppercase tracking-wider font-semibold self-start sm:self-auto"
        >
          Status: Open
        </span>
      </div>

      <!-- Primary Page Header -->
      <div class="flex flex-col gap-1 mb-space-xl">
        <h1 class="font-headline-xl text-headline-xl text-on-surface font-bold tracking-tight">
          Create Ticket
        </h1>
      </div>

      <BaseAlert v-if="error" variant="error" :message="error" class="mb-space-lg" />

      <!-- Master Form Container -->
      <form class="flex flex-col gap-space-xl" @submit.prevent="handleSubmit">
        <div class="grid grid-cols-1 lg:grid-cols-12 gap-space-xl items-start">
          <!-- Left Column (Core Spec) - 8 cols -->
          <div class="lg:col-span-8 flex flex-col gap-space-lg">
            <div class="bg-surface-container-lowest rounded-xl p-space-xl shadow-sm flex flex-col gap-space-lg">
              <!-- Title -->
              <div class="flex flex-col gap-space-sm">
                <div class="flex items-center justify-between">
                  <label
                    for="ticket-title"
                    class="font-label-md text-label-md text-on-surface font-semibold flex items-center gap-1"
                  >
                    Title <span class="text-error">*</span>
                  </label>
                  <span class="font-code-sm text-code-sm text-outline">
                    {{ title.length }} / 250
                  </span>
                </div>
                <input
                  id="ticket-title"
                  v-model="title"
                  type="text"
                  maxlength="250"
                  placeholder="e.g., Kafka consumer disconnect during ingest spikes"
                  required
                  class="w-full h-10 px-3 rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md placeholder:text-outline focus:outline-none focus:bg-surface-container-lowest focus:shadow-[0_0_0_2px_rgba(79,70,229,0.25)] transition-all"
                />
              </div>

              <!-- Description & Steps -->
              <MarkdownEditor id="ticket-desc" v-model="description" required />
            </div>

            <!-- Supporting Artifacts Block -->
            <div class="bg-surface-container-lowest rounded-xl p-space-lg shadow-sm flex flex-col gap-space-md">
              <div class="flex items-center justify-between">
                <div class="flex items-center gap-space-sm text-on-surface">
                  <BaseIcon name="attach_file" class="text-primary text-[20px]" />
                  <div>
                    <span class="font-label-md text-label-md font-semibold">Supporting Artifacts</span>
                    <p class="font-body-sm text-body-sm text-outline">
                      Attach up to 5 images (PNG, JPG, WebP, GIF, max 5 MB each)
                    </p>
                  </div>
                </div>
                <input
                  ref="fileInputRef"
                  type="file"
                  accept="image/png,image/jpeg,image/webp,image/gif"
                  multiple
                  class="hidden"
                  @change="handleFilesSelected"
                />
                <button
                  type="button"
                  :disabled="pendingImages.length >= 5"
                  class="px-3 py-1.5 rounded-lg bg-surface-container text-on-surface font-label-md text-label-md hover:bg-surface-container-high transition-colors flex items-center gap-1.5 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
                  @click="fileInputRef?.click()"
                >
                  <BaseIcon name="upload" class="text-[18px]" />
                  <span>Upload Images</span>
                </button>
              </div>

              <!-- Live Image Previews -->
              <div v-if="pendingImages.length > 0" class="flex flex-wrap gap-space-sm">
                <div
                  v-for="(item, idx) in pendingImages"
                  :key="item.previewUrl"
                  class="relative w-28 h-24 rounded-lg border border-outline-variant/60 bg-surface-container-low overflow-hidden group"
                >
                  <img
                    :src="item.previewUrl"
                    :alt="item.file.name"
                    class="w-full h-full object-cover cursor-pointer"
                    @click="previewImageUrl = item.previewUrl"
                  />
                  <button
                    type="button"
                    class="absolute top-1 right-1 w-6 h-6 rounded-full bg-on-surface/75 text-surface flex items-center justify-center hover:bg-error transition-colors cursor-pointer"
                    aria-label="Remove image"
                    @click="removePendingImage(idx)"
                  >
                    <BaseIcon name="close" class="text-[14px]" />
                  </button>
                  <div
                    class="absolute inset-x-0 bottom-0 px-1.5 py-0.5 bg-on-surface/70 text-surface text-[10px] truncate pointer-events-none"
                  >
                    {{ item.file.name }}
                  </div>
                </div>
              </div>
            </div>

            <!-- Live Lightbox Preview Modal -->
            <Teleport to="body">
              <div
                v-if="previewImageUrl"
                class="fixed inset-0 z-[70] flex items-center justify-center bg-on-surface/80 backdrop-blur-xs p-space-lg"
                role="dialog"
                aria-modal="true"
                @click.self="previewImageUrl = null"
              >
                <div class="relative max-w-3xl max-h-[88vh] rounded-xl bg-surface-container-lowest p-space-sm shadow-2xl">
                  <button
                    type="button"
                    class="absolute top-3 right-3 w-8 h-8 rounded-full bg-on-surface/75 text-surface flex items-center justify-center hover:bg-on-surface cursor-pointer"
                    aria-label="Close preview"
                    @click="previewImageUrl = null"
                  >
                    <BaseIcon name="close" class="text-[18px]" />
                  </button>
                  <img
                    :src="previewImageUrl"
                    alt="Attachment preview"
                    class="max-h-[82vh] max-w-full object-contain rounded-lg"
                  />
                </div>
              </div>
            </Teleport>
          </div>

          <!-- Right Column (Categorization, SLA, Assignment) - 4 cols -->
          <div class="lg:col-span-4 flex flex-col gap-space-lg">
            <div class="bg-surface-container-lowest rounded-xl p-space-xl shadow-sm flex flex-col gap-space-xl">
              <!-- Category -->
              <div class="flex flex-col gap-space-sm">
                <label
                  for="ticket-category"
                  class="font-label-md text-label-md text-on-surface font-semibold flex items-center gap-1"
                >
                  Category <span class="text-error">*</span>
                </label>
                <div class="relative">
                  <select
                    id="ticket-category"
                    v-model="category"
                    required
                    class="w-full h-10 pl-3 pr-8 rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md appearance-none focus:outline-none focus:bg-surface-container-lowest focus:shadow-[0_0_0_2px_rgba(79,70,229,0.25)] transition-all cursor-pointer"
                  >
                    <option disabled value="">Select category...</option>
                    <option v-for="cat in categories" :key="cat" :value="cat">
                      {{ cat }}
                    </option>
                  </select>
                  <BaseIcon
                    name="unfold_more"
                    class="absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none text-outline text-[18px]"
                  />
                </div>
              </div>

              <!-- Priority -->
              <PrioritySelector v-model="priority" />

              <!-- Deadline -->
              <div class="flex flex-col gap-space-sm">
                <label
                  for="ticket-deadline"
                  class="font-label-md text-label-md text-on-surface font-semibold flex items-center gap-1"
                >
                  Deadline <span class="text-error">*</span>
                </label>
                <input
                  id="ticket-deadline"
                  v-model="deadlineLocal"
                  type="datetime-local"
                  :min="minDeadlineLocal"
                  required
                  class="w-full h-10 px-3 rounded-lg bg-surface-container-low text-on-surface font-code-sm text-code-sm focus:outline-none focus:bg-surface-container-lowest focus:shadow-[0_0_0_2px_rgba(79,70,229,0.25)] transition-all"
                />
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
              to="/tickets"
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
              <span>{{ loading ? 'Creating...' : 'Create Ticket' }}</span>
              <span v-if="!loading" class="font-bold text-[15px]">→</span>
            </button>
          </div>
        </div>
      </form>
    </div>
  </DashboardLayout>
</template>
