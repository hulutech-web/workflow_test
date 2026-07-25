<template>
    <div class="p-4">
        <a-card title="流程设计">
            <template #extra>
                <a-button type="primary" @click="toAdd">新建流程</a-button>
            </template>
            <vxe-grid ref='xGrid' v-bind="gridOptions" v-on="gridEvent">
                <template #publish="{ row }">
                    <span>{{ row.is_publish == true ? "已发布" : "未发布" }}</span>
                </template>
                <template #show="{ row }">
                    <span>{{ row.show == true ? "显示" : "隐藏" }}</span>
                </template>
                <template #design="{ row }">
                    <a-button type="primary" @click="toDesign(row.id)">
                        <PartitionOutlined />管理流程图
                    </a-button>
                </template>
                <template #action="{ row }">
                    <a-button-group>
                        <a-button type="primary" @click="editFlow(row)">编辑</a-button>
                        <a-popconfirm title="确定要删除该流程吗？" ok-text="删除" cancel-text="取消"
                            ok-type="danger" @confirm="deleteFlow(row)">
                            <a-button type="primary" danger>删除</a-button>
                        </a-popconfirm>
                        <a-button type="primary" @click="startFlow(row)" :disabled="row.is_publish == false">
                            <FormatPainterOutlined />发起流程
                        </a-button>
                        <a-button type="primary" @click="startPlugin(row)">
                            <ApiOutlined /> 插件功能
                        </a-button>
                    </a-button-group>
                </template>
                <template #dept="{ row }">
                    <span>{{ row.Dept.id == 0 ? "未分配" : row.Dept.dept_name }}</span>
                </template>
            </vxe-grid>
        </a-card>
    </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
const { gridOptions, destroyFlow } = useFlow()
const router = useRouter();

const xGrid = ref()
const gridEvent: VxeGridListeners<RowVO> = {
    proxyQuery() {
        console.log('数据代理查询事件')
    },
    proxyDelete() {
        console.log('数据代理删除事件')
    },
    proxySave() {
        console.log('数据代理保存事件')
    }
}

const toDesign = (id) => {
    router.push({ path: `/admin/base/flow/${id}/design` })
}

const toAdd = () => {
    router.push({ path: `/admin/base/flow/create` })
}

const startFlow = (row) => {
    if (row.is_publish == false) {
        message.error("流程尚未发布，无法发起流程")
        return
    }
    router.push({ path: `/admin/base/flow/${row.id}/initiation` })
}


const editFlow = (row) => {
    router.push({ path: `/admin/base/flow/${row.id}/edit` })
}

const deleteFlow = async (row: any) => {
    try {
        await destroyFlow(row.id)
        message.success('删除成功')
        xGrid.value?.commitProxy('query')
    } catch (e) {
        // error handled by interceptor
    }
}

const startPlugin = (row) => {
    router.push({ path: `/admin/base/plugin/index`, query: { id: row.id } })
}
</script>

<style scoped></style>
