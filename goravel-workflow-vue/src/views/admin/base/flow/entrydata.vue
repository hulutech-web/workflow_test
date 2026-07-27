<template>
    <div class="p-2">
        <div class="mb-2">
            <a-button type="link" size="small" @click="goBack">
                <ArrowLeftOutlined /> 返回
            </a-button>
        </div>
        <div class="flex justify-center">
            <div style="max-width: 750px; width: 100%">
                <!-- Entry Header -->
                <a-card size="small" class="mb-3">
                    <template #title>
                        <div class="text-center">
                            <span class="text-base font-semibold">{{ flowName }}</span>
                            <a-tag v-if="entry.status === 0" color="processing" class="ml-2">处理中</a-tag>
                            <a-tag v-if="entry.status === 9" color="success" class="ml-2">已完成</a-tag>
                            <a-tag v-if="entry.status === -1" color="error" class="ml-2">已驳回</a-tag>
                            <a-tag v-if="entry.status === -2" color="error" class="ml-2">已撤销</a-tag>
                        </div>
                    </template>
                    <a-descriptions :column="2" size="small">
                        <a-descriptions-item label="发起人">{{ getEmpName() }}</a-descriptions-item>
                        <a-descriptions-item label="当前环节">{{ getProcessName() }}</a-descriptions-item>
                        <a-descriptions-item label="发起时间">{{ entry.created_at }}</a-descriptions-item>
                        <a-descriptions-item label="圆次">{{ entry.circle ?? '-' }}</a-descriptions-item>
                    </a-descriptions>
                </a-card>

                <!-- Form Data -->
                <a-card size="small" title="表单数据" class="mb-3">
                    <a-alert v-if="canEdit" message="该流程已被驳回或撤回，修改后可重新提交" type="warning" show-icon size="small" class="mb-2" />

                    <template v-if="canEdit">
                        <a-form ref="formRef" :model="formState" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }"
                            size="small" :rules="rules">
                            <Form :fields="fillFields" :entryDatas="entryDatas" ref="huluFormRef">
                                <template #default></template>
                            </Form>
                        </a-form>
                        <div class="text-center mt-4">
                            <a-button type="primary" @click="onSubmit">重新提交</a-button>
                        </div>
                    </template>
                    <template v-else>
                        <a-descriptions v-if="fillFields.length" :column="1" size="small" bordered>
                            <a-descriptions-item
                                v-for="f in fillFields"
                                :key="f.field"
                                :label="f.field_name"
                            >
                                {{ entryDataMap[f.field] || '-' }}
                            </a-descriptions-item>
                        </a-descriptions>
                        <a-empty v-else description="暂无表单数据" />
                    </template>
                </a-card>

                <!-- Approval Timeline -->
                <a-card v-if="procs.length > 0" size="small" title="审批记录" class="mb-3">
                    <a-timeline>
                        <a-timeline-item
                            v-for="p in procs"
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

                <!-- Child Workflow Approval Timelines (recursive) -->
                <template v-if="children.length > 0">
                    <ChildApprovalCard
                        v-for="child in children"
                        :key="child.entry?.id || child.entry?.ID"
                        :child-data="child"
                    />
                </template>

                <!-- CC Records -->
                <a-card v-if="ccRecords.length > 0" size="small" title="抄送记录" class="mb-3">
                    <a-tag v-for="cc in ccRecords" :key="cc.id" class="mr-2 mb-1" :color="cc.status == 1 ? 'green' : 'blue'">
                        {{ cc.emp_name }}
                        <span class="ml-1 text-xs opacity-70">({{ cc.status == 1 ? '已读' : '未读' }})</span>
                    </a-tag>
                </a-card>

                <!-- Comments -->
                <a-card v-if="comments.length > 0" size="small" title="评论记录" class="mb-3">
                    <CommentThread
                        :comments="comments"
                        :entry-id="entry.id"
                        :is-readonly="true"
                    />
                </a-card>

                <!-- Empty state when no data loaded yet -->
                <a-empty v-if="!entry.id" description="加载中..." />
            </div>
        </div>
    </div>
</template>

<script setup lang='ts'>
import { message } from 'ant-design-vue'
import Form from '@/components/form/index.vue'
import ChildApprovalCard from "@/views/admin/base/flow/ChildApprovalCard.vue";
import CommentThread from '@/components/comment/CommentThread.vue';

const { showEntry, updateEntry } = useEntry()
const route = useRoute()
const router = useRouter()
const goBack = () => router.back()

const flow_id = route.params.flow_id as string
const entry_id = route.params.entry_id as string

const entry = ref<any>({})
const entryDatas = ref<any[]>([])
const fillFields = ref<any[]>([])
const procs = ref<any[]>([])
const comments = ref<any[]>([])
const ccRecords = ref<any[]>([])
const children = ref<any[]>([])

const flowName = computed(() => {
    return entry.value?.Flow?.flow_name || entry.value?.flow?.flow_name || '-'
})

const canEdit = computed(() => entry.value.status === -1 || entry.value.status === -2)

const entryDataMap = computed(() => {
    const map: Record<string, string> = {}
    entryDatas.value.forEach((ed: any) => {
        map[ed.field_name] = ed.field_value
    })
    return map
})

const getEmpName = () => {
    return entry.value?.Emp?.name || entry.value?.emp?.name || '-'
}
const getProcessName = () => {
    return entry.value?.Process?.process_name || entry.value?.process?.process_name || '-'
}

const init = async () => {
    try {
        const { data } = await showEntry(entry_id)
        if (data.entry) {
            // New Show API returns { entry, comments }
            entry.value = data.entry
            entryDatas.value = data.entry.EntryDatas || []
            fillFields.value = data.entry.Flow?.Template?.TemplateForms || []
            procs.value = data.entry.Procs || []
            comments.value = data.comments || []
            ccRecords.value = data.cc_records || []
            children.value = data.children || []
        } else {
            // Old format: entry object directly
            entry.value = data
            entryDatas.value = data.EntryDatas || []
            fillFields.value = data.Flow?.Template?.TemplateForms || []
            procs.value = data.Procs || []
            comments.value = []
            ccRecords.value = data.cc_records || []
        }

        // Populate form state for edit mode
        if (canEdit.value && entryDatas.value.length) {
            const obj: any = {}
            entryDatas.value.forEach((ed: any) => {
                obj[ed.field_name] = ed.field_value
            })
            formState.value = obj
        }
    } catch (e) {
        // handled
    }
}

// Edit mode form
const formRef = ref()
const huluFormRef = ref()
const formState = ref<any>({})

const rules = computed(() => {
    const r: any = {}
    fillFields.value.forEach((f: any) => {
        if (f.required) {
            r[f.field] = [{ required: true, message: `请输入${f.field_name || f.label}`, trigger: 'blur' }]
        }
    })
    return r
})

const onSubmit = async () => {
    try {
        await updateEntry(entry_id, { ...formState.value, flow_id: +flow_id })
        message.success('重新提交成功')
        goBack()
    } catch (e) {
        // handled by interceptor
    }
}

// Timeline helpers
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

init()
</script>

<style scoped></style>
