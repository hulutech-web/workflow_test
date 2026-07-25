<template>
    <div class="p-4">
        <a-card title="流程实例管理">
            <div class="mb-3">
                <a-space>
                    <a-select v-model:value="statusFilter" placeholder="状态筛选" allow-clear style="width: 150px"
                        @change="handleFilterChange">
                        <a-select-option :value="0">处理中</a-select-option>
                        <a-select-option :value="9">已完成</a-select-option>
                        <a-select-option :value="-1">已驳回</a-select-option>
                        <a-select-option :value="-2">已撤销</a-select-option>
                    </a-select>
                    <a-input-search v-model:value="searchTitle" placeholder="搜索标题" allow-clear style="width: 200px"
                        @search="handleFilterChange" />
                    <a-button @click="handleFilterChange">查询</a-button>
                    <a-button type="primary" @click="refresh">刷新</a-button>
                </a-space>
            </div>
            <vxe-grid ref="xGrid" v-bind="gridOptions" v-on="gridEvent">
                <template #status="{ row }">
                    <a-badge v-if="row.status == 0" status="processing" text="处理中" />
                    <a-badge v-if="row.status == 9" status="success" text="已完成" />
                    <a-badge v-if="row.status == -1" status="error" text="已驳回" />
                    <a-badge v-if="row.status == -2" status="error" text="已撤销" />
                </template>
                <template #action="{ row }">
                    <a-button-group>
                        <a-button type="link" @click="viewEntry(row)">查看</a-button>
                        <a-button type="link" @click="viewEntryData(row)">表单数据</a-button>
                        <a-button type="link" @click="viewProcs(row)">进程明细</a-button>
                        <a-button v-if="row.status === -1 || row.status === -2" type="link" @click="editEntryData(row)" danger>编辑重发</a-button>
                        <a-button v-if="row.status === 0 || row.status === -1" type="link" @click="resendEntryFn(row)"
                            danger>重发</a-button>
                        <a-button v-if="row.status === 0" type="link" @click="revokeEntryFn(row)" danger>撤回</a-button>
                    </a-button-group>
                </template>
            </vxe-grid>
        </a-card>

        <!-- Entry Detail Modal -->
        <a-modal :footer="false" v-model:open="detailOpen" title="实例详情" centered width="800px">
            <div v-if="currentEntry">
                <a-descriptions bordered :column="2">
                    <a-descriptions-item label="标题">{{ currentEntry.title }}</a-descriptions-item>
                    <a-descriptions-item label="流程">{{ currentEntry.Flow?.flow_name }}</a-descriptions-item>
                    <a-descriptions-item label="发起人">{{ currentEntry.Emp?.name }}</a-descriptions-item>
                    <a-descriptions-item label="当前环节">{{ currentEntry.Process?.process_name }}</a-descriptions-item>
                    <a-descriptions-item label="状态">
                        <a-badge v-if="currentEntry.status == 0" status="processing" text="处理中" />
                        <a-badge v-if="currentEntry.status == 9" status="success" text="已完成" />
                        <a-badge v-if="currentEntry.status == -1" status="error" text="已驳回" />
                        <a-badge v-if="currentEntry.status == -2" status="error" text="已撤销" />
                    </a-descriptions-item>
                    <a-descriptions-item label="发起时间">{{ currentEntry.created_at }}</a-descriptions-item>
                    <a-descriptions-item label="圆次">{{ currentEntry.circle }}</a-descriptions-item>
                    <a-descriptions-item label="父实例">{{ currentEntry.pid > 0 ? currentEntry.pid : '-' }}</a-descriptions-item>
                </a-descriptions>
            </div>
        </a-modal>

        <!-- Procs Modal -->
        <a-modal :footer="false" v-model:open="procsOpen" title="进程明细" centered width="900px">
            <vxe-grid v-bind="procsGridOptions" :data="currentProcs" />
        </a-modal>
    </div>
</template>

<script setup lang="ts">
import { message, Modal } from 'ant-design-vue'

const { showEntry, resendEntry, revokeEntry } = useEntry()
const { indexProcs } = useProc()
const { getCcList } = useCc()
const router = useRouter()

const xGrid = ref()
const statusFilter = ref<number | undefined>(undefined)
const searchTitle = ref('')

const gridOptions = {
    border: true,
    stripe: true,
    showOverflow: true,
    columns: [
        { field: 'id', title: 'ID', width: 80 },
        { field: 'title', title: '标题', minWidth: 200 },
        { field: 'flow_name', title: '流程名称', minWidth: 150 },
        { field: 'emp_name', title: '发起人', width: 100 },
        { field: 'process_name', title: '当前环节', width: 120 },
        { title: '状态', slotName: 'status', width: 100 },
        { field: 'created_at', title: '创建时间', width: 180 },
        { title: '操作', slotName: 'action', width: 250 },
    ],
    proxyConfig: {
        proxy: true,
        ajax: {
            query: async ({ page }) => {
                // Use home API to get entries since backend Index returns nil
                const { data } = await http.request({ url: 'home', method: 'GET' })
                let entries = data?.entries || []

                if (statusFilter.value !== undefined) {
                    entries = entries.filter((e: any) => e.status === statusFilter.value)
                }
                if (searchTitle.value) {
                    entries = entries.filter((e: any) => e.title?.includes(searchTitle.value))
                }

                return {
                    total: entries.length,
                    items: entries.slice((page.currentPage - 1) * page.pageSize, page.currentPage * page.pageSize)
                }
            }
        }
    }
}

const gridEvent: VxeGridListeners<RowVO> = {
    proxyQuery() {
        console.log('Entry list loaded')
    }
}

const handleFilterChange = () => {
    xGrid.value?.commitProxy('query')
}

const refresh = () => {
    statusFilter.value = undefined
    searchTitle.value = ''
    handleFilterChange()
}

// Detail
const detailOpen = ref(false)
const currentEntry = ref<any>(null)

const viewEntry = async (row: any) => {
    try {
        const { data } = await showEntry(row.id)
        currentEntry.value = data
        detailOpen.value = true
    } catch (e) {
        currentEntry.value = row
        detailOpen.value = true
    }
}

const viewEntryData = (row: any) => {
    router.push({ path: `/admin/base/flow/${row.flow_id}/entry/${row.id}` })
}

const editEntryData = (row: any) => {
    router.push({ path: `/admin/base/flow/${row.flow_id}/entry/${row.id}` })
}

const revokeEntryFn = async (row: any) => {
    Modal.confirm({
        title: '确认撤回',
        content: `确定要撤回「${row.title}」吗？此操作不可撤销。`,
        okText: '确认撤回',
        cancelText: '取消',
        okType: 'danger',
        onOk: async () => {
            try {
                await revokeEntry(row.id)
                message.success('撤回成功')
                handleFilterChange()
            } catch (e) {
                // handled by interceptor
            }
        }
    })
}

const resendEntryFn = async (row: any) => {
    try {
        await resendEntry(row.id)
        message.success('重发成功')
        handleFilterChange()
    } catch (e) {
        // handled by interceptor
    }
}

// Procs modal
const procsOpen = ref(false)
const currentProcs = ref<any[]>([])

const procsGridOptions = {
    border: true,
    stripe: true,
    showOverflow: true,
    columns: [
        { field: 'id', title: 'ID', width: 60 },
        { field: 'process_name', title: '环节', width: 120 },
        { field: 'emp_name', title: '审批人', width: 100 },
        { field: 'content', title: '意见', minWidth: 200 },
        { title: '状态', slotName: 'proc_status', width: 100 },
        { field: 'created_at', title: '时间', width: 180 },
    ],
}

const viewProcs = async (row: any) => {
    try {
        const { data } = await indexProcs(row.id)
        currentProcs.value = data || []
        procsOpen.value = true
    } catch (e) {
        currentProcs.value = []
        procsOpen.value = true
    }
}

onMounted(() => {
    xGrid.value?.commitProxy('query')
})
</script>

<style scoped></style>
