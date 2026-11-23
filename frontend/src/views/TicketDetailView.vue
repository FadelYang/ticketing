<script setup lang="ts">
import { computed, onMounted, ref, watch } from 'vue'
import { useRoute } from 'vue-router'
import { storeToRefs } from 'pinia'
import { ApiError } from '@/services/api'
import { attachmentService } from '@/services/attachmentService'
import { ticketService } from '@/services/ticketService'
import { useTicketSocket } from '@/composables/useTicketSocket'
import { useAuthStore } from '@/stores/authStore'
import { useTicketStore } from '@/stores/ticketStore'
import DashboardLayout from '@/components/layout/DashboardLayout.vue'
import MarkdownEditor from '@/components/tickets/MarkdownEditor.vue'
import PrioritySelector from '@/components/tickets/PrioritySelector.vue'
import TicketComments from '@/components/tickets/TicketComments.vue'
import AttachmentGallery from '@/components/tickets/AttachmentGallery.vue'
import BaseAlert from '@/components/common/BaseAlert.vue'
import BaseIcon from '@/components/common/BaseIcon.vue'
import type { Attachment } from '@/types/attachment'
import type {
  CreateCommentPayload,
  Ticket,
  TicketComment,
  TicketPriority,
  TicketStatus,
} from '@/types/ticket'

const route = useRoute()
const authStore = useAuthStore()
const { user } = storeToRefs(authStore)

const ticketStore = useTicketStore()
const { total, assignees, userMap } = storeToRefs(ticketStore)
const { lastTicketEvent } = useTicketSocket()

watch(lastTicketEvent, (event) => {
  if (!event || !ticket.value) return
  if (event.type === 'ticket.updated' && event.data.id === ticket.value.id) {
    ticket.value = event.data
    selectedAssigneeId.value = event.data.assignee_id ?? ''
  }
})

const ticket = ref<Ticket | null>(null)
const selectedAssigneeId = ref('')
const loading = ref(true)
const updating = ref(false)
const error = ref<string | null>(null)
const actionError = ref<string | null>(null)
const actionSuccess = ref<string | null>(null)

const comments = ref<TicketComment[]>([])
const commentsTotal = ref(0)
const commentsPage = ref(1)
const commentsTotalPages = ref(1)
const commentsSort = ref<'oldest' | 'newest'>('oldest')
const commentsLoading = ref(false)
const commentsSubmitting = ref(false)
const commentsError = ref<string | null>(null)
const commentSubmitError = ref<string | null>(null)
const commentsRef = ref<InstanceType<typeof TicketComments> | null>(null)

const commentDraft = ref('')
const ticketAttachments = ref<Attachment[]>([])
const commentAttachmentsByCommentId = ref<Record<string, Attachment[]>>({})

const isAdmin = computed(() => user.value?.role === 'admin')
const canEditPriority = computed(
  () => user.value?.role === 'admin' || user.value?.role === 'agent',
)

const statusBadgeMap: Record<TicketStatus, { label: string; className: string }> = {
  open: {
    label: 'Status: Open',
    className: 'bg-secondary-fixed text-on-secondary-fixed',
  },
  in_progress: {
    label: 'Status: In Progress',
    className: 'bg-primary-fixed text-on-primary-fixed-variant',
  },
  resolved: {
    label: 'Status: Resolved',
    className: 'bg-surface-container text-on-surface',
  },
  closed: {
    label: 'Status: Closed',
    className: 'bg-surface-container-high text-outline',
  },
}

const statusBadge = computed(() => {
  if (!ticket.value) return statusBadgeMap.open
  return statusBadgeMap[ticket.value.status] ?? statusBadgeMap.open
})

const formattedDeadline = computed(() => {
  if (!ticket.value?.deadline_at) return ''
  const date = new Date(ticket.value.deadline_at)
  if (Number.isNaN(date.getTime())) return ''
  return new Date(date.getTime() - date.getTimezoneOffset() * 60000)
    .toISOString()
    .slice(0, 16)
})

const currentAssigneeLabel = computed(() => {
  const id = ticket.value?.assignee_id
  if (!id) return 'Unassigned'
  const found = userMap.value.get(id)
  return found ? `${found.name} (${found.email})` : id
})

interface StatusAction {
  status: TicketStatus
  label: string
  variantClass: string
}

// Mirrors backend validateStatusTransition rules in backend/internal/module/ticket/usecase.go
const allowedStatusActions = computed<StatusAction[]>(() => {
  const current = ticket.value
  const currentUser = user.value
  if (!current || !currentUser) return []

  const role = currentUser.role
  const isAssignedAgent = role === 'agent' && current.assignee_id === currentUser.id
  const isTicketRequester = role === 'requester' && current.requester_id === currentUser.id

  const actions: StatusAction[] = []

  if (current.status === 'open') {
    if (role === 'admin' || isAssignedAgent) {
      actions.push({
        status: 'in_progress',
        label: 'Start Progress',
        variantClass: 'bg-primary-container hover:bg-primary text-on-primary',
      })
    }
    if (role === 'admin' || isTicketRequester) {
      actions.push({
        status: 'closed',
        label: 'Close Ticket',
        variantClass: 'bg-surface-container-high hover:bg-surface-container-highest text-on-surface',
      })
    }
  } else if (current.status === 'in_progress') {
    if (role === 'admin' || isAssignedAgent) {
      actions.push({
        status: 'resolved',
        label: 'Mark Resolved',
        variantClass: 'bg-secondary hover:opacity-95 text-on-secondary',
      })
    }
  } else if (current.status === 'resolved') {
    if (isTicketRequester) {
      actions.push({
        status: 'open',
        label: 'Reopen Ticket',
        variantClass: 'bg-surface-container-high hover:bg-surface-container-highest text-on-surface',
      })
    }
    if (role === 'admin' || isAssignedAgent || isTicketRequester) {
      actions.push({
        status: 'closed',
        label: 'Close Ticket',
        variantClass: 'bg-primary-container hover:bg-primary text-on-primary',
      })
    }
  } else if (current.status === 'closed') {
    if (isTicketRequester) {
      actions.push({
        status: 'open',
        label: 'Reopen Ticket',
        variantClass: 'bg-surface-container-high hover:bg-surface-container-highest text-on-surface',
      })
    }
  }

  return actions
})

onMounted(async () => {
  const id = String(route.params.id || '')
  if (!id) {
    error.value = 'Ticket ID is missing.'
    loading.value = false
    return
  }

  await ticketStore.fetchUsers()

  try {
    const loaded = await ticketService.getById(id)
    ticket.value = loaded
    selectedAssigneeId.value = loaded.assignee_id ?? ''
    await Promise.all([fetchTicketAttachments(id), fetchComments(id, 1, false)])
  } catch (err) {
    error.value = err instanceof ApiError ? err.message : 'Failed to load ticket details.'
  } finally {
    loading.value = false
  }
})

async function fetchTicketAttachments(ticketId: string) {
  try {
    ticketAttachments.value = await attachmentService.list('ticket', [ticketId])
  } catch {
    ticketAttachments.value = []
  }
}

async function fetchCommentAttachments(commentIds: string[]) {
  if (commentIds.length === 0) return
  try {
    const items = await attachmentService.list('comment', commentIds)
    const grouped: Record<string, Attachment[]> = { ...commentAttachmentsByCommentId.value }
    for (const id of commentIds) {
      grouped[id] = []
    }
    for (const att of items) {
      if (!grouped[att.target_id]) grouped[att.target_id] = []
      grouped[att.target_id].push(att)
    }
    commentAttachmentsByCommentId.value = grouped
  } catch {
    // Non-blocking
  }
}

async function fetchComments(ticketId: string, page = 1, append = false) {
  commentsLoading.value = true
  commentsError.value = null

  try {
    const res = await ticketService.getComments(ticketId, {
      sort: commentsSort.value,
      page,
      limit: 20,
    })
    const items = res.items ?? []
    comments.value = append ? [...comments.value, ...items] : items
    commentsTotal.value = res.total
    commentsPage.value = res.page
    commentsTotalPages.value = Math.max(1, res.total_pages)
    await fetchCommentAttachments(items.map((item) => item.id))
  } catch (err) {
    commentsError.value = err instanceof ApiError ? err.message : 'Failed to load comments.'
  } finally {
    commentsLoading.value = false
  }
}

async function handleCommentSortChange(nextSort: 'oldest' | 'newest') {
  if (!ticket.value || commentsSort.value === nextSort) return
  commentsSort.value = nextSort
  await fetchComments(ticket.value.id, 1, false)
}

async function handleLoadMoreComments() {
  if (!ticket.value || commentsPage.value >= commentsTotalPages.value) return
  await fetchComments(ticket.value.id, commentsPage.value + 1, true)
}

async function handleCommentSubmit(payload: CreateCommentPayload, files: File[] = []) {
  if (!ticket.value) return
  commentsSubmitting.value = true
  commentSubmitError.value = null

  try {
    const created = await ticketService.createComment(ticket.value.id, payload)
    if (files.length > 0) {
      const uploaded = await attachmentService.upload('comment', created.id, files)
      commentAttachmentsByCommentId.value = {
        ...commentAttachmentsByCommentId.value,
        [created.id]: uploaded,
      }
    }
    if (commentsSort.value === 'newest') {
      comments.value = [created, ...comments.value]
    } else if (commentsPage.value >= commentsTotalPages.value) {
      comments.value = [...comments.value, created]
    } else {
      await fetchComments(ticket.value.id, 1, false)
    }
    commentsTotal.value += 1
    commentsRef.value?.resetComposer()
  } catch (err) {
    commentSubmitError.value = err instanceof ApiError ? err.message : 'Failed to post comment.'
  } finally {
    commentsSubmitting.value = false
  }
}

async function handleAssign() {
  if (!ticket.value || !isAdmin.value) return
  updating.value = true
  actionError.value = null
  actionSuccess.value = null

  try {
    const updated = await ticketService.patch(ticket.value.id, {
      assignee_id: selectedAssigneeId.value ? selectedAssigneeId.value : null,
    })
    ticket.value = updated
    selectedAssigneeId.value = updated.assignee_id ?? ''
    actionSuccess.value = updated.assignee_id
      ? 'Assignee updated successfully.'
      : 'Ticket unassigned.'
  } catch (err) {
    actionError.value = err instanceof ApiError ? err.message : 'Failed to update assignee.'
  } finally {
    updating.value = false
  }
}

async function handlePriorityChange(nextPriority: TicketPriority) {
  if (!ticket.value || !canEditPriority.value || nextPriority === ticket.value.priority) return
  updating.value = true
  actionError.value = null
  actionSuccess.value = null

  try {
    const updated = await ticketService.patch(ticket.value.id, {
      priority: nextPriority,
    })
    ticket.value = updated
    actionSuccess.value = `Priority updated to ${updated.priority}.`
  } catch (err) {
    actionError.value = err instanceof ApiError ? err.message : 'Failed to update priority.'
  } finally {
    updating.value = false
  }
}

async function handleStatusChange(nextStatus: TicketStatus) {
  if (!ticket.value) return
  const isClosing = nextStatus === 'closed'
  const isReopening =
    nextStatus === 'open' &&
    (ticket.value.status === 'resolved' || ticket.value.status === 'closed')

  const trimmedComment = commentDraft.value.trim()

  if ((isClosing || isReopening) && !trimmedComment) {
    const msg = isClosing
      ? 'Please fill in the comment form before closing the ticket.'
      : 'Please fill in the comment form before reopening the ticket.'
    actionError.value = msg
    commentSubmitError.value = msg
    commentsRef.value?.focusComposer()
    return
  }

  updating.value = true
  actionError.value = null
  actionSuccess.value = null
  commentSubmitError.value = null

  try {
    const updated = await ticketService.patch(ticket.value.id, {
      status: nextStatus,
      ...(isClosing || isReopening ? { comment: trimmedComment } : {}),
    })
    ticket.value = updated
    if (isClosing || isReopening) {
      commentsRef.value?.resetComposer()
      await fetchComments(ticket.value.id, 1, false)
    }
    actionSuccess.value = `Status updated to ${statusBadgeMap[updated.status].label.replace('Status: ', '')}.`
  } catch (err) {
    actionError.value = err instanceof ApiError ? err.message : 'Failed to update status.'
  } finally {
    updating.value = false
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
            TICKET DETAIL
          </span>
        </div>

        <span
          v-if="ticket"
          class="px-2.5 py-0.5 rounded-full font-label-sm text-label-sm uppercase tracking-wider font-semibold self-start sm:self-auto"
          :class="statusBadge.className"
        >
          {{ statusBadge.label }}
        </span>
      </div>

      <!-- Primary Page Header -->
      <div class="flex flex-col gap-1 mb-space-xl">
        <h1 class="font-headline-xl text-headline-xl text-on-surface font-bold tracking-tight">
          Ticket Details
        </h1>
      </div>

      <div
        v-if="loading"
        class="bg-surface-container-lowest rounded-xl p-space-2xl shadow-sm flex items-center justify-center gap-2 text-outline"
      >
        <BaseIcon name="progress_activity" class="text-lg animate-spin" />
        <span class="font-body-md text-body-md">Loading ticket details...</span>
      </div>

      <BaseAlert v-else-if="error" variant="error" :message="error" />

      <!-- Read-only Form Container (with Admin Assign & Role-based Status Controls) -->
      <form v-else-if="ticket" class="flex flex-col gap-space-lg" @submit.prevent>
        <BaseAlert v-if="actionError" variant="error" :message="actionError" />
        <BaseAlert v-if="actionSuccess" variant="success" :message="actionSuccess" />

        <div class="grid grid-cols-1 lg:grid-cols-12 gap-space-xl items-start">
          <!-- Left Column (Core Spec) - 8 cols -->
          <div class="lg:col-span-8 flex flex-col gap-space-lg">
            <div
              class="bg-surface-container-lowest rounded-xl p-space-xl shadow-sm flex flex-col gap-space-lg"
            >
              <!-- Title -->
              <div class="flex flex-col gap-space-sm">
                <div class="flex items-center justify-between">
                  <label
                    for="detail-title"
                    class="font-label-md text-label-md text-on-surface font-semibold"
                  >
                    Title
                  </label>
                  <span class="font-code-sm text-code-sm text-outline">
                    {{ ticket.title.length }} / 250
                  </span>
                </div>
                <input
                  id="detail-title"
                  :value="ticket.title"
                  type="text"
                  disabled
                  class="w-full h-10 px-3 rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md disabled:opacity-80 disabled:cursor-not-allowed"
                />
              </div>

              <!-- Description & Steps -->
              <MarkdownEditor id="detail-desc" :model-value="ticket.description" disabled />

              <!-- Ticket Attachments -->
              <div v-if="ticketAttachments.length > 0" class="flex flex-col gap-space-xs pt-space-sm border-t border-outline-variant/40">
                <span class="font-label-md text-label-md text-on-surface font-semibold">
                  Attachments ({{ ticketAttachments.length }})
                </span>
                <AttachmentGallery :attachments="ticketAttachments" />
              </div>
            </div>
          </div>

          <!-- Right Column (Categorization, Priority, Assignee, Deadline) - 4 cols -->
          <div class="lg:col-span-4 flex flex-col gap-space-lg">
            <div
              class="bg-surface-container-lowest rounded-xl p-space-xl shadow-sm flex flex-col gap-space-xl"
            >
              <!-- Category -->
              <div class="flex flex-col gap-space-sm">
                <label
                  for="detail-category"
                  class="font-label-md text-label-md text-on-surface font-semibold"
                >
                  Category
                </label>
                <div class="relative">
                  <select
                    id="detail-category"
                    :value="ticket.category"
                    disabled
                    class="w-full h-10 pl-3 pr-8 rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md appearance-none disabled:opacity-80 disabled:cursor-not-allowed"
                  >
                    <option :value="ticket.category">{{ ticket.category }}</option>
                  </select>
                  <BaseIcon
                    name="unfold_more"
                    class="absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none text-outline text-[18px]"
                  />
                </div>
              </div>

              <!-- Priority -->
              <PrioritySelector
                :model-value="ticket.priority"
                :disabled="!canEditPriority || updating"
                @update:model-value="handlePriorityChange"
              />

              <!-- Assignee (Editable for Admin, Read-only for Agent/Requester) -->
              <div class="flex flex-col gap-space-sm">
                <label
                  for="detail-assignee"
                  class="font-label-md text-label-md text-on-surface font-semibold"
                >
                  Assignee
                </label>

                <template v-if="isAdmin">
                  <div class="flex items-center gap-2">
                    <div class="relative flex-1">
                      <select
                        id="detail-assignee"
                        v-model="selectedAssigneeId"
                        :disabled="updating"
                        class="w-full h-10 pl-3 pr-8 rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md appearance-none focus:outline-none focus:bg-surface-container-lowest focus:shadow-[0_0_0_2px_rgba(79,70,229,0.25)] transition-all cursor-pointer disabled:opacity-60"
                      >
                        <option value="">Unassigned</option>
                        <option v-for="agent in assignees" :key="agent.id" :value="agent.id">
                          {{ agent.name }} ({{ agent.role }})
                        </option>
                      </select>
                      <BaseIcon
                        name="person_search"
                        class="absolute right-2.5 top-1/2 -translate-y-1/2 pointer-events-none text-outline text-[18px]"
                      />
                    </div>
                    <button
                      type="button"
                      :disabled="updating"
                      class="h-10 px-3.5 rounded-lg bg-primary-container hover:bg-primary text-on-primary font-label-md text-label-md font-semibold shadow-sm transition-all cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed shrink-0"
                      @click="handleAssign"
                    >
                      Assign
                    </button>
                  </div>
                </template>

                <input
                  v-else
                  id="detail-assignee"
                  :value="currentAssigneeLabel"
                  type="text"
                  disabled
                  class="w-full h-10 px-3 rounded-lg bg-surface-container-low text-on-surface font-body-md text-body-md disabled:opacity-80 disabled:cursor-not-allowed"
                />
              </div>

              <!-- Deadline -->
              <div class="flex flex-col gap-space-sm">
                <label
                  for="detail-deadline"
                  class="font-label-md text-label-md text-on-surface font-semibold"
                >
                  Deadline
                </label>
                <input
                  id="detail-deadline"
                  :value="formattedDeadline"
                  type="datetime-local"
                  disabled
                  class="w-full h-10 px-3 rounded-lg bg-surface-container-low text-on-surface font-code-sm text-code-sm disabled:opacity-80 disabled:cursor-not-allowed"
                />
              </div>
            </div>
          </div>
        </div>

        <!-- Sticky Bottom Action Footer -->
        <div
          class="sticky bottom-0 z-30 -mx-gutter-desktop px-gutter-desktop py-space-md bg-surface-container-lowest/90 backdrop-blur-md shadow-md rounded-t-xl flex flex-col sm:flex-row items-center justify-between gap-space-md mt-space-lg"
        >
          <RouterLink
            to="/tickets"
            class="inline-flex items-center gap-2 px-4 py-2 rounded-lg bg-surface-container hover:bg-surface-container-high text-on-surface font-label-md text-label-md font-semibold transition-all"
          >
            <BaseIcon name="arrow_back" class="text-[18px]" />
            <span>Back to Tickets</span>
          </RouterLink>

          <div
            v-if="allowedStatusActions.length > 0"
            class="flex flex-wrap items-center gap-space-sm justify-end"
          >
            <button
              v-for="action in allowedStatusActions"
              :key="action.status"
              type="button"
              :disabled="updating"
              class="inline-flex items-center gap-2 px-4 py-2 rounded-lg font-label-md text-label-md font-semibold shadow-sm transition-all cursor-pointer disabled:opacity-60 disabled:cursor-not-allowed"
              :class="action.variantClass"
              @click="handleStatusChange(action.status)"
            >
              <BaseIcon v-if="updating" name="progress_activity" class="text-[16px] animate-spin" />
              <span>{{ action.label }}</span>
            </button>
          </div>
        </div>
      </form>

      <!-- PR-Style Ticket Comments Below Detail Ticket -->
      <TicketComments
        v-if="ticket"
        ref="commentsRef"
        v-model:draft="commentDraft"
        :comments="comments"
        :total="commentsTotal"
        :loading="commentsLoading"
        :submitting="commentsSubmitting"
        :error="commentsError"
        :submit-error="commentSubmitError"
        :sort="commentsSort"
        :has-more="comments.length < commentsTotal"
        :requester-id="ticket.requester_id"
        :user-map="userMap"
        :current-user="user"
        :attachments-by-comment="commentAttachmentsByCommentId"
        @submit="handleCommentSubmit"
        @update:sort="handleCommentSortChange"
        @load-more="handleLoadMoreComments"
      />
    </div>
  </DashboardLayout>
</template>
