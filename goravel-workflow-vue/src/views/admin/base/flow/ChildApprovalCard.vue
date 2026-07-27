<template>
    <a-card size="small" :title="'子流程 — ' + flowName" class="mb-3" :bordered="true"
        :style="{ borderLeft: '3px solid #1890ff' }">
        <!-- Child Entry Header -->
        <a-descriptions :column="2" size="small" class="mb-2">
            <a-descriptions-item label="发起人">{{ childEntry?.Emp?.name || childEntry?.emp?.name || '-' }}</a-descriptions-item>
            <a-descriptions-item label="当前环节">{{ childEntry?.Process?.process_name || childEntry?.process?.process_name || '-' }}</a-descriptions-item>
            <a-descriptions-item label="状态">
                <a-tag v-if="childEntry?.status === 0" color="processing">处理中</a-tag>
                <a-tag v-if="childEntry?.status === 9" color="success">已完成</a-tag>
                <a-tag v-if="childEntry?.status === -1" color="error">已驳回</a-tag>
                <a-tag v-if="childEntry?.status === -2" color="error">已撤销</a-tag>
            </a-descriptions-item>
        </a-descriptions>

        <!-- Child Form Data -->
        <a-card v-if="formFields.length" size="small" title="表单数据" class="mb-2">
            <a-descriptions :column="1" size="small" bordered>
                <a-descriptions-item v-for="f in formFields" :key="f.field" :label="f.field_name">
                    {{ entryDataMap[f.field] || '-' }}
                </a-descriptions-item>
            </a-descriptions>
        </a-card>

        <!-- Child Procs Timeline -->
        <a-card v-if="childProcs.length > 0" size="small" title="审批记录" class="mb-2">
            <a-timeline>
                <a-timeline-item
                    v-for="p in childProcs"
                    :key="p.id"
                    :color="timelineColor(p.status)"
                >
                    <div class="flex items-center gap-2 flex-wrap">
                        <span class="font-semibold">{{ p.process_name }}</span>
                        <span class="text-xs text-gray-500">— {{ p.emp_name || p.auditor_name }}</span>
                        <a-tag :color="procStatusColor(p.status)" size="small">{{ procStatusText(p.status) }}</a-tag>
                    </div>
                    <div v-if="p.content" class="text-sm text-gray-600 mt-1">
                        <span class="text-gray-500">意见：</span>{{ p.content }}
                    </div>
                    <div class="text-xs text-gray-400 mt-1">{{ p.created_at }}</div>
                </a-timeline-item>
            </a-timeline>
        </a-card>

        <!-- Child Comments -->
        <a-card v-if="childComments.length > 0" size="small" title="评论记录" class="mb-2">
            <a-list item-layout="horizontal" :data-source="childComments" size="small">
                <template #renderItem="{ item }">
                    <a-list-item>
                        <a-list-item-meta>
                            <template #title>{{ item.emp_name }}</template>
                            <template #description>{{ item.created_at }}</template>
                        </a-list-item-meta>
                        <span>{{ item.content }}</span>
                    </a-list-item>
                </template>
            </a-list>
        </a-card>

        <!-- Recursive grandchildren -->
        <template v-if="grandChildren.length > 0">
            <ChildApprovalCard
                v-for="gc in grandChildren"
                :key="gc.entry?.id || gc.entry?.ID"
                :child-data="gc"
            />
        </template>
    </a-card>
</template>

<script setup lang="ts">
const props = defineProps<{
    childData: any
}>()

const childEntry = computed(() => props.childData?.entry || {})
const childProcs = computed(() => childEntry.value?.Procs || [])
const childComments = computed(() => props.childData?.comments || [])
const grandChildren = computed(() => props.childData?.children || [])

const flowName = computed(() => {
    return childEntry.value?.Flow?.flow_name || childEntry.value?.flow?.flow_name || '-'
})

const formFields = computed(() => {
    return childEntry.value?.Flow?.Template?.TemplateForms || childEntry.value?.flow?.Template?.TemplateForms || []
})

const entryDatas = computed(() => childEntry.value?.EntryDatas || [])

const entryDataMap = computed(() => {
    const map: Record<string, string> = {}
    entryDatas.value.forEach((ed: any) => {
        map[ed.field_name] = ed.field_value
    })
    return map
})

// Timeline helpers (same as parent)
const timelineColor = (status: number) => {
    if (status === 1 || status === 9) return 'green'
    if (status === -1 || status === -2) return 'red'
    if (status === 3 || status === 4) return 'gray'
    return 'blue'
}
const procStatusText = (status: number) => {
    if (status === 0) return '处理中'
    if (status === 1 || status === 9) return '已通过'
    if (status === -1) return '已驳回'
    if (status === -2) return '已撤回'
    if (status === 3) return '已转交'
    if (status === 4) return '已跳过'
    return '未知'
}
const procStatusColor = (status: number) => {
    if (status === 0) return 'processing'
    if (status === 1 || status === 9) return 'success'
    if (status === -1 || status === -2) return 'error'
    return 'default'
}
</script>

<script lang="ts">
export default {
    name: 'ChildApprovalCard',
}
</script>
