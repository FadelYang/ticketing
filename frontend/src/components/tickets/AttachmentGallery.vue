<script setup lang="ts">
import { onUnmounted, ref, watch } from 'vue'
import { attachmentService } from '@/services/attachmentService'
import BaseIcon from '@/components/common/BaseIcon.vue'
import type { Attachment } from '@/types/attachment'

interface Props {
  attachments: Attachment[]
}

const props = defineProps<Props>()

const blobUrls = ref<Record<string, string>>({})
const previewIndex = ref<number | null>(null)

function formatSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

async function loadImages(items: Attachment[]) {
  for (const item of items) {
    if (blobUrls.value[item.id]) continue
    try {
      const url = await attachmentService.fetchFileBlobUrl(item.id)
      blobUrls.value = { ...blobUrls.value, [item.id]: url }
    } catch {
      // Skip broken image gracefully
    }
  }
}

watch(
  () => props.attachments,
  (items) => {
    void loadImages(items)
  },
  { immediate: true },
)

onUnmounted(() => {
  for (const url of Object.values(blobUrls.value)) {
    URL.revokeObjectURL(url)
  }
})

function openPreview(index: number) {
  previewIndex.value = index
}

function closePreview() {
  previewIndex.value = null
}

function prevImage() {
  if (previewIndex.value === null || props.attachments.length === 0) return
  previewIndex.value = (previewIndex.value - 1 + props.attachments.length) % props.attachments.length
}

function nextImage() {
  if (previewIndex.value === null || props.attachments.length === 0) return
  previewIndex.value = (previewIndex.value + 1) % props.attachments.length
}
</script>

<template>
  <div v-if="attachments.length > 0" class="flex flex-wrap gap-space-sm pt-space-sm">
    <button
      v-for="(att, idx) in attachments"
      :key="att.id"
      type="button"
      class="group relative w-28 h-24 rounded-lg border border-outline-variant/60 bg-surface-container-low overflow-hidden shadow-xs hover:border-primary transition-all cursor-pointer text-left"
      :title="`${att.file_name} (${formatSize(att.size_bytes)})`"
      @click="openPreview(idx)"
    >
      <img
        v-if="blobUrls[att.id]"
        :src="blobUrls[att.id]"
        :alt="att.file_name"
        class="w-full h-full object-cover group-hover:scale-105 transition-transform duration-150"
      />
      <div
        v-else
        class="w-full h-full flex items-center justify-center text-outline"
      >
        <BaseIcon name="image" class="text-[22px] animate-pulse" />
      </div>
      <div
        class="absolute inset-x-0 bottom-0 px-1.5 py-0.5 bg-on-surface/70 text-surface text-[10px] truncate"
      >
        {{ att.file_name }}
      </div>
    </button>

    <!-- Live Lightbox Preview Modal -->
    <Teleport to="body">
      <div
        v-if="previewIndex !== null && attachments[previewIndex]"
        class="fixed inset-0 z-[70] flex items-center justify-center bg-on-surface/80 backdrop-blur-xs p-space-lg"
        role="dialog"
        aria-modal="true"
        @click.self="closePreview"
      >
        <div
          class="relative max-w-4xl w-full max-h-[90vh] flex flex-col rounded-xl bg-surface-container-lowest border border-outline-variant/60 shadow-2xl overflow-hidden"
        >
          <div
            class="px-space-lg py-space-sm bg-surface-container-low border-b border-outline-variant/50 flex items-center justify-between gap-2"
          >
            <div class="flex items-center gap-2 min-w-0">
              <BaseIcon name="image" class="text-[18px] text-primary shrink-0" />
              <span class="font-label-md text-label-md font-semibold text-on-surface truncate">
                {{ attachments[previewIndex].file_name }}
              </span>
              <span class="font-code-sm text-code-sm text-outline shrink-0">
                ({{ formatSize(attachments[previewIndex].size_bytes) }})
              </span>
            </div>
            <button
              type="button"
              class="p-1 rounded-lg text-outline hover:text-on-surface hover:bg-surface-container transition-colors cursor-pointer"
              aria-label="Close preview"
              @click="closePreview"
            >
              <BaseIcon name="close" class="text-[20px]" />
            </button>
          </div>

          <div class="relative flex-1 bg-surface flex items-center justify-center p-space-md overflow-auto">
            <img
              v-if="blobUrls[attachments[previewIndex].id]"
              :src="blobUrls[attachments[previewIndex].id]"
              :alt="attachments[previewIndex].file_name"
              class="max-h-[72vh] max-w-full object-contain rounded-lg"
            />

            <button
              v-if="attachments.length > 1"
              type="button"
              class="absolute left-3 top-1/2 -translate-y-1/2 w-9 h-9 rounded-full bg-surface-container-lowest/90 shadow-md flex items-center justify-center text-on-surface hover:bg-surface-container cursor-pointer"
              aria-label="Previous image"
              @click="prevImage"
            >
              <BaseIcon name="chevron_left" class="text-[22px]" />
            </button>
            <button
              v-if="attachments.length > 1"
              type="button"
              class="absolute right-3 top-1/2 -translate-y-1/2 w-9 h-9 rounded-full bg-surface-container-lowest/90 shadow-md flex items-center justify-center text-on-surface hover:bg-surface-container cursor-pointer"
              aria-label="Next image"
              @click="nextImage"
            >
              <BaseIcon name="chevron_right" class="text-[22px]" />
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
