<template>
    <div class="p-2">
        <a-card size="small" title="员工管理">
            <template #extra>
                <a-button type="primary" size="small" @click="addEmp">创建员工</a-button>
            </template>
            <vxe-grid ref='xGrid' v-bind="{ ...gridOptions, size: 'small' }" v-on="gridEvent">
                <template #action="{ row }">
                    <a-space size="small">
                        <a-button type="primary" size="small">删除</a-button>
                        <a-button type="primary" size="small">编辑</a-button>
                        <a-button type="primary" size="small" @click="bind(row)">绑定用户</a-button>
                    </a-space>
                </template>
                <template #dept="{ row }">
                    <span>{{ row.Dept.id == 0 ? "未分配" : row.Dept.dept_name }}</span>
                </template>
            </vxe-grid>
            <a-modal :footer="false" v-model:open="open" width="800px" title="用户" centered size="small"
                :bodyStyle="{ height: '600px' }">
                <Userlist @bind="bindins" />
            </a-modal>
        </a-card>
    </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue';


const { gridOptions, bindUser } = useEmp()
const router = useRouter()
const rulesStore = useRulesStore()

// TABLE
const xGrid = ref()
const gridEvent: VxeGridListeners<RowVO> = {
    proxyQuery() {
        const grid = xGrid.value
        const data = grid.getTableData().fullData
    },
    proxyDelete() {},
    proxySave() {}
}
const open = ref(false)
const addEmp = () => {
    router.push({ path: "/admin/manage/groupcourse/crt" })
}
const formRef = ref()

const state = ref({
    user_id: null,
    emp_id: null
})
const bindins = async (val) => {
    state.value.user_id = val.user_id
    if (state.value.user_id == null || state.value.emp_id == null) {
        return message.error("为选中用户")
    }
    await bindUser(state.value)
}


const onSubmit = async () => {
    try {
        formRef.value.clearValidate()
    } catch (error) {
        // handled
    }
}


const bind = (val) => {
    open.value = true
    state.value.emp_id = val.id
}
</script>

<style scoped></style>
