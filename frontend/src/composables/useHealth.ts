import { onMounted, ref } from 'vue'
import { exampleService } from '@/services/exampleService'

export function useHealth() {
  const status = ref<'checking' | 'healthy' | 'offline'>('checking')
  const database = ref<string>('unknown')

  async function refresh() {
    status.value = 'checking'
    try {
      const data = await exampleService.checkHealth()
      status.value = data.status === 'ok' ? 'healthy' : 'offline'
      database.value = data.database
    } catch {
      status.value = 'offline'
      database.value = 'disconnected'
    }
  }

  onMounted(refresh)

  return {
    status,
    database,
    refresh,
  }
}
