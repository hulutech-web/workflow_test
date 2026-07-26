<template>
  <div class="p-2">
    <a-page-header title="归档详情" @back="() => router.back()">
      <template #tags>
        <a-tag v-if="archive.status == 9" color="success">已完成</a-tag>
        <a-tag v-if="archive.status == -1" color="error">已驳回</a-tag>
        <a-tag v-if="archive.status == -2" color="error">已撤回</a-tag>
      </template>
    </a-page-header>

    <a-row :gutter="16">
      <a-col :span="16">
        <!-- Form Data -->
        <a-card size="small" title="表单数据" class="mb-3">
          <a-descriptions v-if="formData.length" :column="2" size="small" bordered>
            <a-descriptions-item v-for="item in formData" :key="item.field_name" :label="item.field_name">
              {{ item.field_value || '-' }}
            </a-descriptions-item>
          </a-descriptions>
          <a-empty v-else description="无表单数据" />
        </a-card>

        <!-- Approval Timeline -->
        <a-card size="small" title="审批记录" class="mb-3">
          <a-timeline v-if="procs.length">
            <a-timeline-item
              v-for="proc in procs"
              :key="proc.id"
              :color="timelineColor(proc.status)"
            >
              <div class="flex items-center gap-2">
                <span class="font-medium">{{ proc.emp_name || proc.auditor_name || '系统' }}</span>
                <a-tag :color="procStatusColor(proc.status)" size="small">
                  {{ procStatusText(proc.status) }}
                </a-tag>
              </div>
              <div class="text-gray-500 text-sm">{{ proc.created_at }}</div>
              <div v-if="proc.content" class="mt-1 text-gray-700">{{ proc.content }}</div>
              <div v-if="proc.dept_name" class="text-gray-400 text-xs">{{ proc.dept_name }}</div>
            </a-timeline-item>
          </a-timeline>
          <a-empty v-else description="无审批记录" />
        </a-card>

        <!-- Comments -->
        <a-card size="small" title="评论记录">
          <a-list v-if="comments.length" :data-source="comments" size="small">
            <template #renderItem="{ item }">
              <a-list-item>
                <a-list-item-meta>
                  <template #title>
                    <span class="font-medium">{{ item.emp_name || '-' }}</span>
                    <span class="text-gray-400 text-xs ml-2">{{ item.created_at }}</span>
                  </template>
                  <template #description>
                    {{ item.content || '-' }}
                  </template>
                </a-list-item-meta>
              </a-list-item>
            </template>
          </a-list>
          <a-empty v-else description="无评论" />
        </a-card>
      </a-col>

      <a-col :span="8">
        <!-- Entry Info -->
        <a-card size="small" title="基本信息" class="mb-3">
          <a-descriptions :column="1" size="small" bordered>
            <a-descriptions-item label="标题">{{ archive.title || '-' }}</a-descriptions-item>
            <a-descriptions-item label="流程">{{ flowInfo?.flow_name || '-' }}</a-descriptions-item>
            <a-descriptions-item label="发起人">{{ entryInfo?.Emp?.name || '-' }}</a-descriptions-item>
            <a-descriptions-item label="部门">{{ entryInfo?.Emp?.Dept?.dept_name || entryInfo?.Emp?.dept_name || '-' }}</a-descriptions-item>
            <a-descriptions-item label="归档时间">{{ archive.created_at }}</a-descriptions-item>
          </a-descriptions>
        </a-card>

        <!-- CC Records -->
        <a-card size="small" title="抄送记录">
          <div v-if="ccRecords.length" class="flex flex-col gap-1">
            <a-tag v-for="cc in ccRecords" :key="cc.id" class="w-fit">
              {{ cc.emp_name || '-' }}
              <span class="text-gray-400 text-xs ml-1">
                {{ cc.status == 1 ? '已读' : '未读' }}
              </span>
            </a-tag>
          </div>
          <a-empty v-else description="无抄送" />
        </a-card>
      </a-col>
    </a-row>
  </div>
</template>

<script setup lang="ts">
const router = useRouter()
const route = useRoute()
const { showArchive } = useEntryArchive()

const archive = ref<any>({})
const formData = ref<any[]>([])
const procs = ref<any[]>([])
const comments = ref<any[]>([])
const ccRecords = ref<any[]>([])
const flowInfo = ref<any>({})
const entryInfo = ref<any>({})

const timelineColor = (status: number) => {
  if (status == 1 || status == 9) return 'green'
  if (status == -1) return 'red'
  if (status == -2 || status == 4) return 'gray'
  return 'blue'
}

const procStatusText = (status: number) => {
  const map: Record<number, string> = {
    0: '待处理',
    1: '已通过',
    9: '会签通过',
    [-1]: '已驳回',
    [-2]: '已撤回',
    3: '已转交',
    4: '已跳过',
  }
  return map[status] || '未知'
}

const procStatusColor = (status: number) => {
  if (status == 1 || status == 9) return 'green'
  if (status == -1) return 'red'
  if (status == 3 || status == 4) return 'default'
  if (status == -2) return 'default'
  return 'processing'
}

const parseJSON = (str: string) => {
  if (!str) return null
  try {
    return JSON.parse(str)
  } catch {
    return null
  }
}

onMounted(async () => {
  const id = route.params.id
  try {
    const { data } = await showArchive(Number(id))
    archive.value = data.data || data
    formData.value = parseJSON(archive.value.form_data_snapshot) || []
    procs.value = parseJSON(archive.value.procs_snapshot) || []
    comments.value = parseJSON(archive.value.comments_snapshot) || []
    ccRecords.value = parseJSON(archive.value.cc_snapshot) || []
    flowInfo.value = parseJSON(archive.value.flow_snapshot) || {}
    entryInfo.value = parseJSON(archive.value.entry_snapshot) || {}
  } catch {
    // handled by interceptor
  }
})
</script>
