<template>
  <div class="p-2">
    <p>流程实例管理</p>

      <vxe-grid ref="xGrid" v-bind="gridOptions" v-on="gridEvent">
        <template #title_default="{ row }">
          <span>{{ row.title || getFlowName(row) || `实例${row.id}` }}</span>
        </template>
        <template #status="{ row }">
          <a-badge v-if="row.status == 0" status="processing" text="处理中"/>
          <a-badge v-if="row.status == 9" status="success" text="已完成"/>
          <a-badge v-if="row.status == -1" status="error" text="已驳回"/>
          <a-badge v-if="row.status == -2" status="error" text="已撤销"/>
        </template>
        <template #flow_name="{ row }">
          <span>{{ getFlowName(row) }}</span>
        </template>
        <template #emp_name="{ row }">
          <span>{{ getEmpName(row) }}</span>
        </template>
        <template #process_name="{ row }">
          <span>{{ getProcessName(row) }}</span>
        </template>
        <template #action="{ row }">
          <a-space size="small">
            <a-button type="primary" ghost size="small" @click="viewDetail(row)">详情</a-button>
            <a-button v-if="row.status === -1 || row.status === -2" type="link" size="small" @click="viewDetail(row)"
                      danger>编辑重发
            </a-button>
            <a-button v-if="row.status === 0 || row.status === -1" type="link" size="small" @click="resendEntryFn(row)"
                      danger>重发
            </a-button>
            <a-button v-if="row.status === 0" type="link" size="small" @click="revokeEntryFn(row)" danger>撤回
            </a-button>
          </a-space>
        </template>
      </vxe-grid>
  </div>
</template>

<script setup lang="ts">
import {Modal} from 'ant-design-vue'
import {http} from "@/plugins/axios";
import XEUtils from "xe-utils";

const {resendEntry, revokeEntry,gridOptions} = useEntry()
const router = useRouter()

const xGrid = ref()
const statusFilter = ref<number | undefined>(undefined)
const searchTitle = ref('')

const getFlowName = (row: any) => {
  return row.Flow?.flow_name || row.flow?.flow_name || '-'
}
const getEmpName = (row: any) => {
  return row.Emp?.name || row.emp?.name || '-'
}
const getProcessName = (row: any) => {
  return row.Process?.process_name || row.process?.process_name || '-'
}



const gridEvent: VxeGridListeners<RowVO> = {
  proxyQuery() {
    // loaded
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

// Detail — navigate to unified detail page
const viewDetail = (row: any) => {
  router.push({path: `/admin/base/flow/${row.flow_id}/entry/${row.id}`})
}

const revokeEntryFn = async (row: any) => {
  Modal.confirm({
    title: '确认撤回',
    content: `确定要撤回「${row.title || getFlowName(row)}」吗？此操作不可撤销。`,
    okText: '确认撤回',
    cancelText: '取消',
    okType: 'danger',
    centered: true,
    onOk: async () => {
      try {
        await revokeEntry(row.id)
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
    handleFilterChange()
  } catch (e) {
    // handled by interceptor
  }
}

onMounted(() => {
  xGrid.value?.commitProxy('query')
})
</script>

<style scoped></style>
