<template>
    <div class="p-2">
        <div class="mb-2">
            <a-button type="link" size="small" @click="goBack">
                <ArrowLeftOutlined /> 返回
            </a-button>
        </div>
        <div class="flex justify-center">
            <div style="max-width: 700px; width: 100%">
                <a-watermark :content="entry.status === -1 || entry.status === -2 ? '已驳回' : '已发起'">
                    <a-card size="small">
                        <template #title>
                            <div class="text-center">
                                <span class="text-base font-semibold">{{ flow.flow_name }}</span>
                                <a-tag v-if="flow.Template" color="blue" class="ml-1">
                                    {{ flow.Template.template_name }}
                                </a-tag>
                            </div>
                        </template>

                        <!-- 表单数据（只读展示） -->
                        <a-descriptions v-if="fillFields.length" :column="1" size="small" bordered class="mb-3">
                            <a-descriptions-item
                                v-for="f in fillFields"
                                :key="f.field"
                                :label="f.field_name"
                            >
                                {{ entryDataMap[f.field] || '-' }}
                            </a-descriptions-item>
                        </a-descriptions>

                        <a-divider class="!my-2" />

                        <a-form ref="formRef" :model="formState" :label-col="{ span: 6 }" :wrapper-col="{ span: 16 }"
                            size="small">
                            <a-form-item label="批复内容">
                                <a-textarea v-model:value="formState.content" :rows="3" placeholder="请填写批复内容" />
                            </a-form-item>
                        </a-form>

                        <div class="mt-3 flex flex-wrap gap-2">
                            <a-space size="small">
                                <a-button type="primary" size="small" @click="pass">同意</a-button>
                                <a-button type="primary" size="small" danger @click="unpass">驳回</a-button>
                                <a-button size="small" @click="showUnpassTo">驳回至节点</a-button>
                                <a-button size="small" @click="showAddSign">加签</a-button>
                                <a-button size="small" @click="showTransfer">转交</a-button>
                                <a-button size="small" @click="showComment">评论</a-button>
                                <a-button size="small" @click="showCc">抄送</a-button>
                            </a-space>
                        </div>
                    </a-card>
                </a-watermark>

                <!-- Comment thread -->
                <a-card v-if="comments.length > 0" title="评论记录" class="mt-2" size="small">
                    <CommentThread
                        :comments="comments"
                        :entry-id="entry_id"
                        :proc-id="proc_id"
                        @comment-added="loadComments"
                    />
                </a-card>
            </div>
        </div>

        <!-- Add Sign Modal -->
        <a-modal v-model:open="addSignOpen" title="加签" centered size="small" @ok="handleAddSign">
            <a-form layout="vertical" size="small">
                <a-form-item label="选择员工">
                    <EmpSearch v-model="addSignEmpId" />
                </a-form-item>
                <a-form-item label="加签类型">
                    <a-radio-group v-model:value="addSignType">
                        <a-radio value="before">前置加签</a-radio>
                        <a-radio value="after">后置加签</a-radio>
                    </a-radio-group>
                </a-form-item>
            </a-form>
        </a-modal>

        <!-- Transfer Modal -->
        <a-modal v-model:open="transferOpen" title="转交" centered size="small" @ok="handleTransfer">
            <a-form layout="vertical" size="small">
                <a-form-item label="转交给">
                    <EmpSearch v-model="transferEmpId" />
                </a-form-item>
            </a-form>
        </a-modal>

        <!-- Comment Modal -->
        <a-modal v-model:open="commentOpen" title="添加评论" centered size="small" @ok="handleComment">
            <a-form layout="vertical" size="small">
                <a-form-item label="评论内容">
                    <a-textarea v-model:value="commentContent" :rows="3" placeholder="请输入评论内容" />
                </a-form-item>
            </a-form>
        </a-modal>

        <!-- UnpassTo Modal -->
        <a-modal v-model:open="unpassToOpen" title="驳回至指定节点" centered size="small" @ok="handleUnpassTo">
            <a-form layout="vertical" size="small">
                <a-form-item label="目标节点">
                    <a-select v-model:value="targetProcessId" placeholder="请选择要驳回到的节点" style="width:100%" size="small">
                        <a-select-option v-for="p in rejectableProcesses" :key="p.id" :value="p.id">
                            {{ p.process_name }} (位置: {{ p.position }})
                        </a-select-option>
                    </a-select>
                </a-form-item>
                <a-form-item label="驳回理由">
                    <a-textarea v-model:value="unpassToContent" :rows="3" placeholder="请输入驳回理由" />
                </a-form-item>
            </a-form>
        </a-modal>

        <!-- CC Modal -->
        <a-modal v-model:open="ccOpen" title="抄送" centered size="small" @ok="handleCc">
            <a-form layout="vertical" size="small">
                <a-form-item label="抄送人">
                    <a-select
                        v-model:value="ccEmpIds"
                        mode="multiple"
                        placeholder="选择抄送人"
                        style="width:100%"
                        size="small"
                    >
                        <a-select-option v-for="e in allEmps" :key="e.id" :value="e.id">
                            {{ e.name }}
                        </a-select-option>
                    </a-select>
                </a-form-item>
            </a-form>
        </a-modal>
    </div>
</template>

<script setup lang='ts'>
import { message } from 'ant-design-vue';
import EmpSearch from '@/components/empsearch/index.vue';
import CommentThread from '@/components/comment/CommentThread.vue';

const { loadFlowEntryConfig, getEntryData } = useEntry();
const { setPass, setUnPass, addSign, transferProc, addComment, getComments, getRejectableProcesses } = useProc();
const { addCc } = useCc();
const { getEmpOpt } = useEmp();
const route = useRoute();
const router = useRouter()
const goBack = () => router.back()

const flow_id = route.params.flow_id
const entry_id = route.params.entry_id;
const query = route.query;
const process_id = query.process_id as string
const proc_id = query.proc_id as string

const fillFields = ref([]);
const flow = ref({})
const entry = ref({})
const comments = ref<any[]>([])
const allEmps = ref<any[]>([])
const formState = ref({ content: '' })

const init = async () => {
    if (flow_id) {
        const { data } = await loadFlowEntryConfig(flow_id);
        flow.value = data
        fillFields.value = data.Template?.TemplateForms || []
    }
    if (entry_id) {
        await loadEntryDatas()
        await loadComments()
    }
    // Load all employees for CC picker
    try {
        const { data } = await getEmpOpt()
        allEmps.value = (data || []).map((e: any) => ({ id: e.id || e.ID, name: e.name }))
    } catch (e) {
        // ignore
    }
}

const entryDatas = ref([])
const entryDataMap = computed(() => {
    const map: Record<string, string> = {}
    entryDatas.value.forEach((ed: any) => {
        map[ed.field_name] = ed.field_value
    })
    return map
})
const loadEntryDatas = async () => {
    const { data } = await getEntryData(entry_id)
    entryDatas.value = data.entrydata
    entry.value = data.entry
}

const loadComments = async () => {
    try {
        const { data } = await getComments(entry_id)
        comments.value = data || []
    } catch (e) {
        comments.value = []
    }
}

const pass = async () => {
    try {
        await setPass({
            content: formState.value.content,
            process_id: process_id,
            entry_id: entry_id,
        })
        history.back()
    } catch (e) {
        // error handled by interceptor
    }
}

const unpass = async () => {
    try {
        await setUnPass({
            content: formState.value.content,
            proc_id: proc_id,
            entry_id: entry_id,
        })
        history.back()
    } catch (e) {
        // error handled by interceptor
    }
}

// Add Sign
const addSignOpen = ref(false)
const addSignEmpId = ref(undefined)
const addSignType = ref('before')
const showAddSign = () => { addSignOpen.value = true }
const handleAddSign = async () => {
    if (!addSignEmpId.value) {
        message.warning('请选择员工')
        return
    }
    try {
        await addSign({
            entry_id: entry_id,
            process_id: process_id,
            sign_emp_id: addSignEmpId.value,
            sign_type: addSignType.value,
        })
        addSignOpen.value = false
    } catch (e) {
        // error handled by interceptor
    }
}

// Transfer
const transferOpen = ref(false)
const transferEmpId = ref(undefined)
const showTransfer = () => { transferOpen.value = true }
const handleTransfer = async () => {
    if (!transferEmpId.value) {
        message.warning('请选择员工')
        return
    }
    try {
        await transferProc({
            entry_id: entry_id,
            proc_id: proc_id,
            target_emp_id: transferEmpId.value,
        })
        transferOpen.value = false
    } catch (e) {
        // error handled by interceptor
    }
}

// Comment
const commentOpen = ref(false)
const commentContent = ref('')
const showComment = () => { commentOpen.value = true }
const handleComment = async () => {
    if (!commentContent.value.trim()) {
        message.warning('请输入评论内容')
        return
    }
    try {
        await addComment({
            entry_id: entry_id,
            content: commentContent.value,
        })
        commentContent.value = ''
        commentOpen.value = false
        await loadComments()
    } catch (e) {
        // error handled by interceptor
    }
}

// UnpassTo
const unpassToOpen = ref(false)
const targetProcessId = ref(undefined)
const unpassToContent = ref('')
const rejectableProcesses = ref<any[]>([])
const showUnpassTo = async () => {
    try {
        const { data } = await getRejectableProcesses(entry_id)
        rejectableProcesses.value = data || []
        targetProcessId.value = undefined
        unpassToContent.value = ''
        unpassToOpen.value = true
    } catch (e) {
        // error handled by interceptor
    }
}
const handleUnpassTo = async () => {
    if (!targetProcessId.value) {
        message.warning('请选择目标节点')
        return
    }
    try {
        await setUnPass({
            proc_id: proc_id,
            entry_id: entry_id,
            content: unpassToContent.value || formState.value.content,
            target_process_id: targetProcessId.value,
        })
        unpassToOpen.value = false
        history.back()
    } catch (e) {
        // error handled by interceptor
    }
}

// CC (抄送)
const ccOpen = ref(false)
const ccEmpIds = ref<number[]>([])
const showCc = () => { ccOpen.value = true; ccEmpIds.value = [] }
const handleCc = async () => {
    if (!ccEmpIds.value.length) {
        message.warning('请选择抄送人')
        return
    }
    try {
        await addCc({
            entry_id: entry_id,
            emp_ids: ccEmpIds.value.join(','),
        })
        ccOpen.value = false
        message.success('抄送成功')
    } catch (e) {
        // error handled by interceptor
    }
}

init();
</script>

<style></style>
