<script setup lang="ts">
import { computed, onUnmounted, ref } from 'vue'
import BaseAlert from '@/components/common/BaseAlert.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'
import AttachmentGallery from '@/components/tickets/AttachmentGallery.vue'
import type { AuthUser } from '@/types/auth'
import type { Attachment } from '@/types/attachment'
import type { CreateCommentPayload, TicketComment } from '@/types/ticket'

interface Props {
  comments: TicketComment[]
  total: number
  loading: boolean
  submitting: boolean
  error: string | null
  submitError: string | null
  sort: 'oldest' | 'newest'
  hasMore: boolean
  requesterId: string
  userMap: Map<string, AuthUser>
  currentUser: AuthUser | null
  draft?: string
  attachmentsByComment?: Record<string, Attachment[]>
}

const props = withDefaults(defineProps<Props>(), {
  draft: '',
  attachmentsByComment: () => ({}),
})

const emit = defineEmits<{
  submit: [payload: CreateCommentPayload, files: File[]]
  'update:sort': [sort: 'oldest' | 'newest']
  'update:draft': [value: string]
  'load-more': []
}>()

const internalDraft = ref('')
const isInternal = ref(false)
const composerTextareaRef = ref<HTMLTextAreaElement | null>(null)
const composerContainerRef = ref<HTMLDivElement | null>(null)
const commentFileInputRef = ref<HTMLInputElement | null>(null)

interface PendingImage {
  file: File
  previewUrl: string
}

const pendingImages = ref<PendingImage[]>([])
const previewImageUrl = ref<string | null>(null)

const draftBody = computed({
  get: () => (props.draft !== undefined ? props.draft : internalDraft.value),
  set: (val: string) => {
    internalDraft.value = val
    emit('update:draft', val)
  },
})

const canCreateInternal = computed(
  () => props.currentUser?.role === 'admin' || props.currentUser?.role === 'agent',
)

const canSubmit = computed(() => draftBody.value.trim().length > 0 && !props.submitting)

function getAuthor(authorId: string): AuthUser | undefined {
  return props.userMap.get(authorId)
}

function getAuthorName(authorId: string): string {
  const author = getAuthor(authorId)
  if (author) return author.name
  if (props.currentUser?.id === authorId) return props.currentUser.name
  return authorId
}

function getAuthorRole(authorId: string): string | null {
  const author = getAuthor(authorId)
  if (author) return author.role
  if (props.currentUser?.id === authorId) return props.currentUser.role
  return null
}

function getInitials(name: string): string {
  const parts = name.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return '?'
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return `${parts[0][0]}${parts[parts.length - 1][0]}`.toUpperCase()
}

function getAvatarClass(role: string | null, internal = false): string {
  if (internal) {
    return 'bg-amber-100 text-amber-800 border-amber-300'
  }
  if (role === 'admin') {
    return 'bg-primary-fixed text-on-primary-fixed-variant border-primary-fixed-dim'
  }
  if (role === 'agent') {
    return 'bg-secondary-fixed text-on-secondary-fixed border-secondary-fixed-dim'
  }
  return 'bg-surface-container-high text-on-surface-variant border-outline-variant'
}

function formatTimestamp(iso: string): string {
  const date = new Date(iso)
  if (Number.isNaN(date.getTime())) return iso
  return new Intl.DateTimeFormat(undefined, {
    month: 'short',
    day: 'numeric',
    year: 'numeric',
    hour: 'numeric',
    minute: '2-digit',
  }).format(date)
}

function handleCommentFilesSelected(event: Event) {
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

function handleSubmit() {
  const body = draftBody.value.trim()
  if (!body || props.submitting) return

  emit(
    'submit',
    {
      body,
      is_internal: canCreateInternal.value ? isInternal.value : false,
    },
    pendingImages.value.map((item) => item.file),
  )
}

function resetComposer() {
  draftBody.value = ''
  isInternal.value = false
  for (const item of pendingImages.value) {
    URL.revokeObjectURL(item.previewUrl)
  }
  pendingImages.value = []
}

onUnmounted(() => {
  for (const item of pendingImages.value) {
    URL.revokeObjectURL(item.previewUrl)
  }
})

function focusComposer() {
  composerContainerRef.value?.scrollIntoView({ behavior: 'smooth', block: 'center' })
  composerTextareaRef.value?.focus()
}

defineExpose({ resetComposer, focusComposer })
</script>

<template>
  <section aria-label="Ticket comments" class="flex flex-col gap-space-lg mt-space-lg">
    <!-- Section Header -->
    <div class="flex flex-wrap items-center justify-between gap-space-sm">
      <div class="flex items-center gap-space-sm">
        <BaseIcon name="forum" class="text-[20px] text-primary" />
        <h2 class="font-headline-md text-headline-md text-on-surface font-semibold">
          Comments
        </h2>
        <span
          class="px-2 py-0.5 rounded-full bg-surface-container-high text-on-surface-variant font-code-sm text-code-sm font-semibold"
        >
          {{ total }}
        </span>
      </div>

      <!-- Sort Toggle -->
      <div
        class="inline-flex items-center rounded-lg bg-surface-container-low p-0.5 border border-outline-variant/40"
      >
        <button
          type="button"
          class="px-3 py-1 rounded-md font-label-sm text-label-sm font-semibold transition-all cursor-pointer"
          :class="
            sort === 'oldest'
              ? 'bg-surface-container-lowest text-on-surface shadow-xs'
              : 'text-outline hover:text-on-surface'
          "
          @click="emit('update:sort', 'oldest')"
        >
          Oldest
        </button>
        <button
          type="button"
          class="px-3 py-1 rounded-md font-label-sm text-label-sm font-semibold transition-all cursor-pointer"
          :class="
            sort === 'newest'
              ? 'bg-surface-container-lowest text-on-surface shadow-xs'
              : 'text-outline hover:text-on-surface'
          "
          @click="emit('update:sort', 'newest')"
        >
          Newest
        </button>
      </div>
    </div>

    <BaseAlert v-if="error" variant="error" :message="error" />

    <!-- Loading State -->
    <div
      v-if="loading && comments.length === 0"
      class="bg-surface-container-lowest rounded-xl p-space-xl shadow-sm flex items-center justify-center gap-2 text-outline"
    >
      <BaseIcon name="progress_activity" class="text-lg animate-spin" />
      <span class="font-body-md text-body-md">Loading comments...</span>
    </div>

    <!-- PR-Style Timeline -->
    <div v-else class="relative flex flex-col gap-space-lg">
      <!-- Vertical timeline connector -->
      <div
        v-if="comments.length > 0"
        aria-hidden="true"
        class="hidden sm:block absolute left-5 top-5 bottom-5 w-0.5 bg-outline-variant/50"
      />

      <!-- Empty State -->
      <div
        v-if="comments.length === 0 && !error"
        class="bg-surface-container-lowest rounded-xl p-space-xl border border-dashed border-outline-variant/70 text-center text-outline font-body-md text-body-md"
      >
        No comments yet. Start the conversation below.
      </div>

      <!-- Comment Items -->
      <article
        v-for="comment in comments"
        :key="comment.id"
        class="relative flex items-start gap-space-md"
      >
        <!-- Author Avatar -->
        <div
          class="hidden sm:flex relative z-10 w-10 h-10 rounded-full border items-center justify-center shrink-0 font-label-sm text-label-sm font-bold shadow-xs"
          :class="getAvatarClass(getAuthorRole(comment.author_id), comment.is_internal)"
          :title="getAuthorName(comment.author_id)"
        >
          {{ getInitials(getAuthorName(comment.author_id)) }}
        </div>

        <!-- PR Comment Card -->
        <div
          class="flex-1 min-w-0 rounded-xl border shadow-sm overflow-hidden transition-colors"
          :class="
            comment.is_internal
              ? 'border-amber-300/80 bg-amber-50/40'
              : 'border-outline-variant/60 bg-surface-container-lowest'
          "
        >
          <!-- PR Comment Header -->
          <header
            class="px-space-lg py-space-sm border-b flex flex-wrap items-center justify-between gap-2"
            :class="
              comment.is_internal
                ? 'bg-amber-100/60 border-amber-300/70'
                : 'bg-surface-container-low border-outline-variant/50'
            "
          >
            <div class="flex flex-wrap items-center gap-1.5 text-body-sm font-body-sm">
              <span class="font-semibold text-on-surface">
                {{ getAuthorName(comment.author_id) }}
              </span>
              <span class="text-outline">commented on</span>
              <time :datetime="comment.created_at" class="text-on-surface-variant font-code-sm text-code-sm">
                {{ formatTimestamp(comment.created_at) }}
              </time>
            </div>

            <div class="flex flex-wrap items-center gap-1.5">
              <span
                v-if="comment.author_id === requesterId"
                class="px-2 py-0.5 rounded-full border border-outline-variant/70 bg-surface-container-lowest text-on-surface-variant font-label-sm text-label-sm font-medium"
              >
                Author
              </span>
              <span
                v-if="getAuthorRole(comment.author_id)"
                class="px-2 py-0.5 rounded-full bg-surface-container-high text-on-surface-variant font-label-sm text-label-sm capitalize"
              >
                {{ getAuthorRole(comment.author_id) }}
              </span>
              <span
                v-if="comment.is_internal"
                class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-200/80 text-amber-900 font-label-sm text-label-sm font-semibold"
              >
                <BaseIcon name="lock" class="text-[13px]" />
                <span>Internal</span>
              </span>
            </div>
          </header>

          <!-- PR Comment Body -->
          <div class="p-space-lg text-body-md font-body-md text-on-surface whitespace-pre-wrap break-words leading-relaxed">
            {{ comment.body }}
            <AttachmentGallery
              v-if="attachmentsByComment[comment.id]?.length"
              :attachments="attachmentsByComment[comment.id]"
            />
          </div>
        </div>
      </article>

      <!-- Load More Button -->
      <div v-if="hasMore" class="flex justify-center pt-space-xs">
        <button
          type="button"
          :disabled="loading"
          class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-surface-container hover:bg-surface-container-high text-on-surface font-label-md text-label-md font-semibold transition-all cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed"
          @click="emit('load-more')"
        >
          <BaseIcon v-if="loading" name="progress_activity" class="text-[16px] animate-spin" />
          <span>Load more comments</span>
        </button>
      </div>

      <!-- PR-Style Comment Composer -->
      <div ref="composerContainerRef" class="relative flex items-start gap-space-md pt-space-xs">
        <div
          class="hidden sm:flex relative z-10 w-10 h-10 rounded-full border items-center justify-center shrink-0 font-label-sm text-label-sm font-bold shadow-xs"
          :class="getAvatarClass(currentUser?.role ?? null, isInternal)"
        >
          {{ getInitials(currentUser?.name || 'You') }}
        </div>

        <form
          class="flex-1 min-w-0 rounded-xl border border-outline-variant/60 bg-surface-container-lowest shadow-sm overflow-hidden"
          @submit.prevent="handleSubmit"
        >
          <div
            class="px-space-lg py-space-sm bg-surface-container-low border-b border-outline-variant/50 flex items-center justify-between"
          >
            <span class="font-label-md text-label-md font-semibold text-on-surface">
              Add a comment
            </span>
            <span
              v-if="isInternal"
              class="inline-flex items-center gap-1 px-2 py-0.5 rounded-full bg-amber-200/80 text-amber-900 font-label-sm text-label-sm font-semibold"
            >
              <BaseIcon name="lock" class="text-[13px]" />
              <span>Internal note</span>
            </span>
          </div>

          <div class="p-space-lg flex flex-col gap-space-md">
            <BaseAlert v-if="submitError" variant="error" :message="submitError" />

            <textarea
              ref="composerTextareaRef"
              v-model="draftBody"
              rows="4"
              placeholder="Leave a comment..."
              :disabled="submitting"
              class="w-full p-space-md rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md placeholder:text-outline focus:outline-none focus:bg-surface-container-lowest focus:shadow-[0_0_0_2px_rgba(79,70,229,0.25)] transition-all resize-y disabled:opacity-60"
            />

            <!-- Live Comment Image Previews -->
            <div v-if="pendingImages.length > 0" class="flex flex-wrap gap-space-sm">
              <div
                v-for="(item, idx) in pendingImages"
                :key="item.previewUrl"
                class="relative w-24 h-20 rounded-lg border border-outline-variant/60 bg-surface-container-low overflow-hidden"
              >
                <img
                  :src="item.previewUrl"
                  :alt="item.file.name"
                  class="w-full h-full object-cover cursor-pointer"
                  @click="previewImageUrl = item.previewUrl"
                />
                <button
                  type="button"
                  class="absolute top-1 right-1 w-5 h-5 rounded-full bg-on-surface/75 text-surface flex items-center justify-center hover:bg-error transition-colors cursor-pointer"
                  aria-label="Remove image"
                  @click="removePendingImage(idx)"
                >
                  <BaseIcon name="close" class="text-[12px]" />
                </button>
              </div>
            </div>

            <div class="flex flex-wrap items-center justify-between gap-space-sm">
              <div class="flex flex-wrap items-center gap-space-md">
                <input
                  ref="commentFileInputRef"
                  type="file"
                  accept="image/png,image/jpeg,image/webp,image/gif"
                  multiple
                  class="hidden"
                  @change="handleCommentFilesSelected"
                />
                <button
                  type="button"
                  :disabled="submitting || pendingImages.length >= 5"
                  class="inline-flex items-center gap-1.5 px-2.5 py-1.5 rounded-lg bg-surface-container hover:bg-surface-container-high text-on-surface font-label-sm text-label-sm font-semibold transition-colors cursor-pointer disabled:opacity-50"
                  @click="commentFileInputRef?.click()"
                >
                  <BaseIcon name="add_photo_alternate" class="text-[16px]" />
                  <span>Attach images</span>
                </button>

                <label
                  v-if="canCreateInternal"
                  class="inline-flex items-center gap-2 text-body-sm font-body-sm text-on-surface-variant cursor-pointer select-none"
                >
                  <input
                    v-model="isInternal"
                    type="checkbox"
                    :disabled="submitting"
                    class="w-4 h-4 rounded border-outline-variant text-primary focus:ring-primary cursor-pointer"
                  />
                  <span>Internal note (visible only to agents &amp; admins)</span>
                </label>
              </div>

              <button
                type="submit"
                :disabled="!canSubmit"
                class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-primary-container hover:bg-primary text-on-primary font-label-md text-label-md font-semibold shadow-sm transition-all cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed"
              >
                <BaseIcon
                  v-if="submitting"
                  name="progress_activity"
                  class="text-[16px] animate-spin"
                />
                <BaseIcon v-else name="send" class="text-[16px]" />
                <span>{{ submitting ? 'Posting...' : 'Comment' }}</span>
              </button>
            </div>
          </div>
        </form>
      </div>
    </div>

    <!-- Live Lightbox Preview Modal for Draft Comment Images -->
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
            alt="Comment attachment preview"
            class="max-h-[82vh] max-w-full object-contain rounded-lg"
          />
        </div>
      </div>
    </Teleport>
  </section>
</template>
