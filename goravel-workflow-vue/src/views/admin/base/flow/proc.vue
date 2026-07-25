<template>
    <div class="p-4">
        <div class="mb-4">
            <a-button type="link" @click="goBack">
                <ArrowLeftOutlined /> 返回
            </a-button>
        </div>
        <div class="flex justify-center">
            <div class="w-full" style="max-width: 700px">
                <a-watermark content="已发起">
                    <a-card>
                        <template #title>
                            <div class="text-center">
                                <span class="text-lg font-semibold">{{ flow.flow_name }}</span>
                                <a-tag v-if="flow.Template" color="blue" class="ml-2">
                                    {{ flow.Template.template_name }}
                                </a-tag>
                            </div>
                        </template>

                        <Form :fields="fillFields" @submit="onSubmit" :entryDatas="entryDatas"
                            ref="huluFormRef">
                            <div>
                                <div class="mb-2 text-base font-medium">批复内容：</div>
                                <a-textarea v-model:value="content" placeholder="请填写批复内容" :rows="4" />
                                <div class="mt-3 flex flex-wrap gap-2">
                                    <a-button type="primary" @click="pass">同意</a-button>
                                    <a-button type="primary" danger @click="unpass">驳回</a-button>
                                    <a-button @click="showUnpassTo">驳回至节点</a-button>
                                    <a-button @click="showAddSign">加签</a-button>
                                    <a-button @click="showTransfer">转交</a-button>
                                    <a-button @click="showComment">评论</a-button>
                                </div>
                            </div>
                        </Form>
                    </a-card>
                </a-watermark>

                <!-- Comment thread -->
                <a-card v-if="comments.length > 0" title="评论记录" class="mt-3">
                    <a-list item-layout="horizontal" :data-source="comments">
                        <template #renderItem="{ item }">
                            <a-list-item>
                                <a-list-item-meta>
                                    <template #title>{{ item.emp_name || '未知用户' }}</template>
                                    <template #description>{{ item.created_at }}</template>
                                </a-list-item-meta>
                                <template #actions>
                                    <span>{{ item.content }}</span>
                                </template>
                            </a-list-item>
                        </template>
                    </a-list>
                </a-card>
            </div>
        </div>

        <!-- Add Sign Modal -->
        <a-modal v-model:open="addSignOpen" title="加签" centered @ok="handleAddSign">
            <a-form layout="vertical">
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
        <a-modal v-model:open="transferOpen" title="转交" centered @ok="handleTransfer">
            <a-form layout="vertical">
                <a-form-item label="转交给">
                    <EmpSearch v-model="transferEmpId" />
                </a-form-item>
            </a-form>
        </a-modal>

        <!-- Comment Modal -->
        <a-modal v-model:open="commentOpen" title="添加评论" centered @ok="handleComment">
            <a-form layout="vertical">
                <a-form-item label="评论内容">
                    <a-textarea v-model:value="commentContent" :rows="4" placeholder="请输入评论内容" />
                </a-form-item>
            </a-form>
        </a-modal>

        <!-- UnpassTo Modal -->
        <a-modal v-model:open="unpassToOpen" title="驳回至指定节点" centered @ok="handleUnpassTo">
            <a-form layout="vertical">
                <a-form-item label="目标节点">
                    <a-select v-model:value="targetProcessId" placeholder="请选择要驳回到的节点" style="width:100%">
                        <a-select-option v-for="p in rejectableProcesses" :key="p.id" :value="p.id">
                            {{ p.process_name }} (位置: {{ p.position }})
                        </a-select-option>
                    </a-select>
                </a-form-item>
                <a-form-item label="驳回理由">
                    <a-textarea v-model:value="unpassToContent" :rows="4" placeholder="请输入驳回理由" />
                </a-form-item>
            </a-form>
        </a-modal>
    </div>
</template>

<script setup lang='ts'>
import { message } from 'ant-design-vue';
import EmpSearch from '@/components/empsearch/index.vue';

const { loadFlowEntryConfig, getEntryData } = useEntry();
const { setPass, setUnPass, addSign, transferProc, addComment, getComments, getRejectableProcesses, indexProcs } = useProc();
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
}

const entryDatas = ref([])
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

const content = ref("")
const huluFormRef = ref()
const onSubmit = async (values) => {
    // handled by action buttons
}

const pass = async () => {
    try {
        const formData = huluFormRef.value?.getFormData?.() || {}
        await setPass({
            ...formData,
            content: content.value,
            process_id: process_id,
            entry_id: entry_id,
        })
        message.success('审批通过')
        history.back()
    } catch (e) {
        // error handled by interceptor
    }
}

const unpass = async () => {
    try {
        await setUnPass({
            content: content.value,
            proc_id: proc_id,
            entry_id: entry_id,
        })
        message.success('已驳回')
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
        message.success('加签成功')
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
        message.success('转交成功')
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
        message.success('评论成功')
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
            content: unpassToContent.value || content.value,
            target_process_id: targetProcessId.value,
        })
        message.success('已驳回至指定节点')
        unpassToOpen.value = false
        history.back()
    } catch (e) {
        // error handled by interceptor
    }
}

init();
</script>

<style></style>
