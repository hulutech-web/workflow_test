<template>
  <div class="p-2">
    <!-- 统计卡片 -->
    <div class="mb-3 grid grid-cols-6 gap-2">
      <div v-for="(item, index) in statCards" :key="index"
           class="shadow hover:shadow-lg hover:scale-105 transition-all duration-300 cursor-pointer flex flex-row items-center justify-center rounded-none p-3 gap-4"
           :style="{ background: item.bg }">
        <component :is="item.icon" style="font-size:32px; color:white;"/>
        <div class="flex space-between items-center">
          <p class="text-base text-white/90">{{ item.label }}</p>
          <p class="text-4xl font-bold text-white">{{ item.value }}</p>
        </div>
      </div>
    </div>

    <!-- 管理面板 -->
    <div class="mb-3">
      <p class="text-xs font-semibold mb-1" style="color:#64748b;">管理面板</p>
      <div class="flex flex-wrap gap-2">
        <div @click="toUser"
             class="shadow hover:shadow-lg hover:scale-105 transition-all duration-300 cursor-pointer flex flex-col items-center justify-center rounded-none p-2 w-24 h-14"
             style="background:linear-gradient(135deg,#3b82f6,#22d3ee);">
          <UserOutlined style="font-size:18px; color:white;"/>
          <p class="text-[10px] text-white font-semibold mt-0.5">用户管理</p>
        </div>
        <div @click="toEmp"
             class="shadow hover:shadow-lg hover:scale-105 transition-all duration-300 cursor-pointer flex flex-col items-center justify-center rounded-none p-2 w-24 h-14"
             style="background:linear-gradient(135deg,#10b981,#14b8a6);">
          <TeamOutlined style="font-size:18px; color:white;"/>
          <p class="text-[10px] text-white font-semibold mt-0.5">员工管理</p>
        </div>
        <div @click="toDept"
             class="shadow hover:shadow-lg hover:scale-105 transition-all duration-300 cursor-pointer flex flex-col items-center justify-center rounded-none p-2 w-24 h-14"
             style="background:linear-gradient(135deg,#f59e0b,#fb923c);">
          <ApartmentOutlined style="font-size:18px; color:white;"/>
          <p class="text-[10px] text-white font-semibold mt-0.5">部门管理</p>
        </div>
        <div @click="toFlow"
             class="shadow hover:shadow-lg hover:scale-105 transition-all duration-300 cursor-pointer flex flex-col items-center justify-center rounded-none p-2 w-24 h-14"
             style="background:linear-gradient(135deg,#f43f5e,#f472b6);">
          <ProjectOutlined style="font-size:18px; color:white;"/>
          <p class="text-[10px] text-white font-semibold mt-0.5">流程管理</p>
        </div>
        <div @click="toTemplate"
             class="shadow hover:shadow-lg hover:scale-105 transition-all duration-300 cursor-pointer flex flex-col items-center justify-center rounded-none p-2 w-24 h-14"
             style="background:linear-gradient(135deg,#8b5cf6,#a78bfa);">
          <FileTextOutlined style="font-size:18px; color:white;"/>
          <p class="text-[10px] text-white font-semibold mt-0.5">模板管理</p>
        </div>
      </div>
    </div>

    <!-- 工作流列表 -->
    <div class="mb-3">
      <p class="text-xs font-semibold mb-1" style="color:#64748b;">工作流</p>
      <div class="flex flex-wrap gap-2">
        <div v-for="(item, index) in state.flows" :key="index"
             class="shadow hover:shadow-lg hover:scale-105 transition-all duration-300 cursor-pointer flex flex-col items-center justify-center rounded-none p-2 w-24 h-14"
             style="background:linear-gradient(135deg,#6366f1,#38bdf8);"
             @click="toWorkflow(item)">
          <ProjectOutlined style="font-size:18px; color:white;"/>
          <p class="text-[10px] text-white font-semibold mt-0.5 text-center leading-tight">{{ item.flow_name }}</p>
        </div>
      </div>
    </div>

    <!-- 待办/已办/抄送 tabs -->
    <div>
      <a-tabs v-model:activeKey="activeKey" class="mt-2">

        <a-tab-pane key="1">
          <template #tab>
            <span>
              待办事项
              <a-badge :count="state.cc_list.length">
              </a-badge>
            </span>
          </template>
          <div class="mb-2 overflow-hidden rounded-none">
            <div class="bg-white rounded-none" style="width:100%;">
              <vxe-table border show-overflow size="small" style="width:100%;"
                         :column-config="{ resizable: true }" :data="state.procs">
                <vxe-column field="title" minWidth="180" title="标题" #default="{ row }">
                  {{ row.Entry?.title || row.Entry?.Flow?.flow_name || '-' }}
                </vxe-column>
                <vxe-column field="user" title="发起人" #default="{ row }">
                  {{ row.Entry?.Emp?.name || '-' }}
                </vxe-column>
                <vxe-column field="process" title="流程位置" #default="{ row }">
                  {{ row.process_name || '-' }}
                </vxe-column>
                <vxe-column field="flow" title="流程" #default="{ row }">
                  {{ row.Flow?.flow_name || '-' }}
                </vxe-column>
                <vxe-column field="auditor" title="审核人" #default="{ row }">
                  {{ row.emp_name || '-' }}
                </vxe-column>
                <vxe-column field="action" title="操作" #default="{ row }">
                  <a-space size="small" wrap>
                    <a-button size="small" type="primary" v-if="row.status == 0" @click="toProc(row)">批复</a-button>
                    <a-button size="small" type="link" v-if="row.status == 9">已处理</a-button>
                    <a-button size="small" type="primary" v-if="row.status == -1">已驳回</a-button>
                    <a-button size="small" ghost @click="toEntryDetail(row)">明细</a-button>
                  </a-space>
                </vxe-column>
              </vxe-table>
            </div>
          </div>
        </a-tab-pane>
        <a-tab-pane key="2">
          <template #tab>
            <span>
              我的申请
              <a-badge :count="state.entries.length">
              </a-badge>
            </span>
          </template>
          <div class="mb-2 overflow-hidden rounded-none">
            <div class="bg-white rounded-none" style="width:100%;">
              <vxe-table border show-overflow ref="tableRef" size="small" style="width:100%;"
                         :column-config="{ resizable: true }" :data="state.entries">
                <vxe-column field="title" minWidth="200" title="标题" #default="{ row }">
                  {{ row.title || row.Flow?.flow_name || '-' }}
                </vxe-column>
                <vxe-column field="emp" title="发起人" #default="{ row }">
                  {{ row.Emp?.name || '-' }}
                </vxe-column>
                <vxe-column field="process" title="当前位置" #default="{ row }">
                  {{ row.Process?.process_name || '-' }}
                </vxe-column>
                <vxe-column field="status" title="状态" #default="{ row }">
                  <a-badge v-if="row.status == 0" status="processing" text="处理中"/>
                  <a-badge v-if="row.status == 9" status="success" text="通过"/>
                  <a-badge v-if="row.status == -1" status="error" text="驳回"/>
                  <a-badge v-if="row.status == -2" status="error" text="撤销"/>
                  <a-badge v-if="row.status == -9" status="warning" text="草稿"/>
                </vxe-column>
                <vxe-column field="action" title="操作" #default="{ row }">
                  <a-space size="small" wrap>
                    <a-button type="primary" size="small" @click="toEntryData(row)">详情
                    </a-button>
                    <a-button v-if="row.status == -1" type="primary" size="small" @click="editEntry(row)">编辑
                    </a-button>
                    <a-button v-if="row.status == -1" size="small" @click="resend(row)">重发</a-button>
                    <a-button v-if="row.status != 9 && row.status != -2" size="small" type="dashed" danger
                              @click="revokeEntryFn(row)">撤销
                    </a-button>
                  </a-space>
                </vxe-column>
              </vxe-table>
            </div>
          </div>
        </a-tab-pane>
        <a-tab-pane key="3">
          <template #tab>
            <span>
              已处理事项
              <a-badge :count="state.handle_procs.length">
              </a-badge>
            </span>
          </template>
          <div class="mb-2 overflow-hidden rounded-none">
            <div class="bg-white rounded-none" style="width:100%;">
              <vxe-table border show-overflow ref="tableRef" size="small" style="width:100%;"
                         :column-config="{ resizable: true }" :data="state.handle_procs">
                <vxe-column field="id" title="id"></vxe-column>
                <vxe-column field="title" minWidth="120" title="标题" #default="{ row }">
                  <span>{{ row.Entry?.title || row.Entry?.Flow?.flow_name || '-' }}</span>
                </vxe-column>
                <vxe-column field="emp" title="发起人" #default="{ row }">
                  <span>{{ row.Emp?.name || '-' }}</span>
                </vxe-column>
                <vxe-column field="process_name" title="环节"></vxe-column>
                <vxe-column field="emp_name" title="当前审批人"></vxe-column>
                <vxe-column field="content" title="意见"></vxe-column>
                <vxe-column field="status" title="状态" #default="{ row }">
                  <a-badge v-if="row.status == 0" status="processing" text="处理中"/>
                  <a-badge v-if="row.status == 9" status="success" text="通过"/>
                  <a-badge v-if="row.status == -1" status="error" text="驳回"/>
                  <a-badge v-if="row.status == -2" status="error" text="撤销"/>
                  <a-badge v-if="row.status == -9" status="warning" text="草稿"/>
                </vxe-column>
                <vxe-column field="dept_name" title="部门"></vxe-column>
                <vxe-column field="detail" title="明细" #default="{ row }">
                  <a-button size="small" @click="toEntryDetail(row)">明细</a-button>
                </vxe-column>
              </vxe-table>
            </div>
          </div>
        </a-tab-pane>
        <a-tab-pane key="4">
          <template #tab>
            <span>
              抄送我的
              <a-badge :count="state.cc_list.length">
              </a-badge>
            </span>
          </template>
          <div class="mb-2 overflow-hidden rounded-none">
            <div class="bg-white rounded-none" style="width:100%;">
              <vxe-table border show-overflow ref="ccTableRef" size="small" style="width:100%;"
                         :column-config="{ resizable: true }" :data="state.cc_list">
                <vxe-column field="entry_title" minWidth="200" title="标题" #default="{ row }">
                  {{ row.Entry?.title || '-' }}
                </vxe-column>
                <vxe-column field="flow_name" title="流程" #default="{ row }">
                  {{ row.Entry?.Flow?.flow_name || '-' }}
                </vxe-column>
                <vxe-column field="sender_name" title="发起人" #default="{ row }">
                  {{ row.Entry?.Emp?.name || '-' }}
                </vxe-column>
                <vxe-column field="process_name" title="环节" #default="{ row }">
                  {{ row.Entry?.Process?.process_name || '-' }}
                </vxe-column>
                <vxe-column field="emp_name" title="抄送人"></vxe-column>
                <vxe-column field="status" title="状态" #default="{ row }">
                  <a-badge v-if="row.status == 0" status="processing" text="未读"/>
                  <a-badge v-if="row.status == 1" status="success" text="已读"/>
                </vxe-column>
                <vxe-column field="created_at" title="抄送时间"></vxe-column>
              </vxe-table>
            </div>
          </div>
        </a-tab-pane>
      </a-tabs>
    </div>

  </div>
</template>
<script lang="ts" setup>
import {ApartmentOutlined, FileTextOutlined, ProjectOutlined, TeamOutlined, UserOutlined} from '@ant-design/icons-vue';

const router = useRouter();
const {index} = useHome()
const {resendEntry, revokeEntry} = useEntry()
const ccTableRef = ref()
const state = ref({
  entries: [],
  flows: [],
  handle_procs: [],
  procs: [],
  cc_list: []
})
const stats = ref({
  dept_count: 0,
  emp_count: 0,
  flow_count: 0,
  template_count: 0,
  entry_count: 0,
  pending_count: 0,
})
const statCards = computed(() => [
  {label: '部门', value: stats.value.dept_count, bg: 'linear-gradient(135deg,#3b82f6,#22d3ee)'},
  {label: '员工', value: stats.value.emp_count, bg: 'linear-gradient(135deg,#10b981,#14b8a6)'},
  {label: '流程', value: stats.value.flow_count, bg: 'linear-gradient(135deg,#f43f5e,#f472b6)'},
  {label: '模板', value: stats.value.template_count, bg: 'linear-gradient(135deg,#8b5cf6,#a78bfa)'},
  {label: '实例', value: stats.value.entry_count, bg: 'linear-gradient(135deg,#f59e0b,#fb923c)'},
  {label: '待办', value: stats.value.pending_count, bg: 'linear-gradient(135deg,#ef4444,#f87171)'},
])
const activeKey = ref('1')
const init = async () => {
  const {data} = await index()
  state.value.entries = data.entries
  state.value.flows = data.flows
  state.value.handle_procs = data.handle_procs
  state.value.procs = data.procs
  state.value.cc_list = data.cc_records || []
  if (data.stats) {
    stats.value.dept_count = data.stats.dept_count ?? 0
    stats.value.emp_count = data.stats.emp_count ?? 0
    stats.value.flow_count = data.stats.flow_count ?? 0
    stats.value.template_count = data.stats.template_count ?? 0
    stats.value.entry_count = data.stats.entry_count ?? 0
    stats.value.pending_count = data.stats.pending_count ?? 0
  }
}
const toEntryDetail = (row: any) => {
  const entryId = row.entry_id || row.id
  const flowId = row.flow_id || row.Flow?.ID || row.Entry?.flow_id
  router.push({ path: `/admin/base/flow/${flowId}/entry/${entryId}` })
}
onMounted(() => {
  init()
})
const toEmp = () => {
  router.push({path: "/admin/base/user/emp"})
}
const toDept = () => {
  router.push({path: "/admin/base/dept/index"})
}

const toFlow = () => {
  router.push({path: "/admin/base/flow/index"})
}
const toUser = () => {
  router.push({path: "/admin/base/user/index"})
}
const toTemplate = () => {
  router.push({path: "/admin/base/template/index"})
}

const toEntryData = (row) => {
  router.push({path: `/admin/base/flow/${row.flow_id}/entry/${row.id}`})
}
const toWorkflow = (row) => {
  router.push({path: `/admin/base/flow/${row.id}/initiation`})
}
const toProc = (row) => {
  router.push({
    path: `/admin/base/flow/${row.flow_id}/proc/${row.entry_id}`,
    query: {process_id: row.process_id, proc_id: row.id}
  })
}

const toPlugin = () => {
  router.push({path: "/admin/base/plugin/index"})
}

const resend = async (row) => {
  await resendEntry(row.id)
  await init();
}

const editEntry = (row) => {
  router.push({path: `/admin/base/flow/${row.flow_id}/entry/${row.id}`})
}

const revokeEntryFn = async (row) => {
  await revokeEntry(row.id)
  await init()
}
</script>

<style scoped></style>
