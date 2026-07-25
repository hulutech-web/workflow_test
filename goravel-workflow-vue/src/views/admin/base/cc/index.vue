<template>
    <div>
        <a-card title="抄送记录">
            <vxe-grid ref="xGrid" v-bind="gridOptions" v-on="gridEvent">
                <template #status="{ row }">
                    <a-badge v-if="row.status == 0" status="processing" text="处理中" />
                    <a-badge v-if="row.status == 9" status="success" text="已完成" />
                    <a-badge v-if="row.status == -1" status="error" text="已驳回" />
                    <a-badge v-if="row.status == -2" status="error" text="已撤销" />
                </template>
                <template #action="{ row }">
                    <a-button type="link" @click="viewCcDetail(row)">详情</a-button>
                </template>
            </vxe-grid>
        </a-card>

        <!-- CC Detail Modal -->
        <a-modal :footer="false" v-model:open="detailOpen" title="抄送详情" centered width="700px">
            <div v-if="currentCc">
                <a-descriptions bordered :column="2">
                    <a-descriptions-item label="标题">{{ currentCc.entry?.title }}</a-descriptions-item>
                    <a-descriptions-item label="流程">{{ currentCc.flow?.flow_name }}</a-descriptions-item>
                    <a-descriptions-item label="环节">{{ currentCc.process_name }}</a-descriptions-item>
                    <a-descriptions-item label="发起人">{{ currentCc.entry?.Emp?.name }}</a-descriptions-item>
                    <a-descriptions-item label="抄送时间">{{ currentCc.created_at }}</a-descriptions-item>
                    <a-descriptions-item label="状态">
                        <a-badge v-if="currentCc.status == 0" status="processing" text="处理中" />
                        <a-badge v-if="currentCc.status == 9" status="success" text="已完成" />
                        <a-badge v-if="currentCc.status == -1" status="error" text="已驳回" />
                    </a-descriptions-item>
                </a-descriptions>
                <p class="mt-3 text-lg font-bold">表单数据</p>
                <a-table :columns="formColumns" :data-source="formFields" bordered size="small">
                    <template #bodyCell="{ column, record }">
                        <span>{{ record.value || '-' }}</span>
                    </template>
                </a-table>
            </div>
        </a-modal>
    </div>
</template>

<script setup lang="ts">
const { getCcList, getEntryCc } = useCc()
const xGrid = ref()

const gridOptions = {
    border: true,
    stripe: true,
    showOverflow: true,
    columns: [
        { field: 'entry_id', title: '申请ID', width: 100 },
        { field: 'entry_title', title: '标题', minWidth: 200 },
        { field: 'flow_name', title: '流程名称', minWidth: 150 },
        { field: 'process_name', title: '环节', width: 120 },
        { field: 'emp_name', title: '发起人', width: 100 },
        { field: 'cc_time', title: '抄送时间', width: 180 },
        { title: '状态', slotName: 'status', width: 100 },
        { title: '操作', slotName: 'action', width: 100 },
    ],
    proxyConfig: {
        proxy: true,
        ajax: {
            query: async ({ page }) => {
                const { data } = await getCcList()
                return {
                    total: data?.length || 0,
                    items: data || []
                }
            }
        }
    }
}

const gridEvent: VxeGridListeners<RowVO> = {
    proxyQuery() {
        const data = xGrid.value?.getTableData()?.fullData
        console.log('CC list loaded:', data)
    }
}

// Detail modal
const detailOpen = ref(false)
const currentCc = ref<any>(null)
const formFields = ref<any[]>([])

const formColumns = [
    { title: '字段名', dataIndex: 'field' },
    { title: '值', dataIndex: 'value' },
]

const viewCcDetail = async (row: any) => {
    currentCc.value = row
    try {
        const { data } = await getEntryCc(row.entry_id)
        formFields.value = data?.entrydata || []
    } catch (e) {
        formFields.value = []
    }
    detailOpen.value = true
}

onMounted(() => {
    xGrid.value?.commitProxy('query')
})
</script>

<style scoped></style>
