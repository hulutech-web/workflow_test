<template>
   <div class="p-4">
      <!-- 标题 -->
      <h2 class="text-xl font-semibold mb-4 m-0">模板管理</h2>

      <!-- 模板列表 + 新建 -->
      <a-card :bordered="false" size="small" style="box-shadow: 0 1px 2px rgba(0,0,0,0.06); margin-bottom: 16px;">
         <template #title>
            <a-button type="primary" size="small" @click="handleNewTemplate">新建模板</a-button>
         </template>
         <vxe-grid ref='xGrid' v-bind="{ ...gridOptions, size: 'small' }" v-on="gridEvent" style="width:100%;">
            <template #action="{ row }">
               <a-space size="small">
                  <a-button size="small" @click="editTemplate(row)">编辑</a-button>
                  <a-popconfirm title="将同步删除模板字段，确认删除吗？" ok-text="是" cancel-text="点错了"
                     @confirm="deleteTemplate(row)">
                     <a-button type="primary" size="small" danger>删除</a-button>
                  </a-popconfirm>
                  <a-button type="primary" size="small" ghost @click="loadTmplForm(row)">管理控件</a-button>
               </a-space>
            </template>
         </vxe-grid>
      </a-card>

      <!-- 控件管理区 -->
      <div v-if="selectedTemplateId" class="bg-white rounded-none p-4" style="box-shadow: 0 1px 2px rgba(0,0,0,0.06);">
         <div class="flex items-center gap-2 mb-3">
            <span class="font-semibold text-sm">控件管理 — {{ selectedTemplateName }}</span>
            <a-button type="primary" size="small" @click="handleAddField">添加控件</a-button>
         </div>

         <vxe-table border show-overflow size="small" style="width:100%;"
            :column-config="{ resizable: true }" :data="fields">
            <vxe-column field="id" title="ID" width="50"></vxe-column>
            <vxe-column field="field_name" title="控件名称" minWidth="120"></vxe-column>
            <vxe-column field="field" title="字段名" minWidth="120"></vxe-column>
            <vxe-column field="field_type" title="类型" width="80">
               <template #default="{ row }">{{ fieldLabel(row.field_type) }}</template>
            </vxe-column>
            <vxe-column field="sort" title="排序" width="60"></vxe-column>
            <vxe-column field="field_value" title="选项" minWidth="150">
               <template #default="{ row }">
                  {{ row.field_value?.join(', ') || '-' }}
               </template>
            </vxe-column>
            <vxe-column field="field_default_value" title="默认值" minWidth="100">
               <template #default="{ row }">{{ row.field_default_value || '-' }}</template>
            </vxe-column>
            <vxe-column field="field_rules" title="验证规则" minWidth="180">
               <template #default="{ row }">
                  <span v-if="!row.field_rules || row.field_rules.length === 0">-</span>
                  <span v-else>
                     <span v-for="(rule, i) in row.field_rules" :key="i" class="mr-2">
                        <a-tag size="small">{{ rule.rule_name }}: {{ rule.rule_title }}</a-tag>
                     </span>
                  </span>
               </template>
            </vxe-column>
            <vxe-column title="操作" width="140">
               <template #default="{ row }">
                  <a-space size="small">
                     <a-button size="small" @click="editField(row)">编辑</a-button>
                     <a-popconfirm title="确定删除?" @confirm="delRecord(row)">
                        <a-button size="small" type="primary" danger>删除</a-button>
                     </a-popconfirm>
                  </a-space>
               </template>
            </vxe-column>
         </vxe-table>
      </div>

      <!-- 无选中提示 -->
      <div v-else class="bg-white rounded-none p-12 text-center" style="box-shadow: 0 1px 2px rgba(0,0,0,0.06);">
         <p class="text-gray-400 text-sm">请从上方列表中选择一个模板以管理控件</p>
      </div>

      <!-- 新建/编辑模板 Modal -->
      <a-modal v-model:open="editOpen" :title="templateUpdateState.id ? '编辑模板' : '新建模板'"
         centered size="small" width="480px" :footer="false">
         <a-form :model="templateUpdateState" ref="templateUptRef" size="small">
            <a-form-item label="模板名称">
               <a-input v-model:value="templateUpdateState.template_name" placeholder="输入模板名称" />
            </a-form-item>
            <a-form-item>
               <a-button type="primary" size="small" @click="submitUpdateTemplate">保存</a-button>
               <a-button size="small" @click="editOpen = false" style="margin-left: 8px;">取消</a-button>
            </a-form-item>
         </a-form>
      </a-modal>

      <!-- 添加/编辑控件 Modal -->
      <a-modal v-model:open="open" :title="cid ? '编辑控件' : '添加控件'" centered size="small" width="600px" :footer="false">
         <Formpart :id="cid" ref="formpartRef" />
      </a-modal>
   </div>
</template>

<script setup lang="ts">
import Formpart from './formpart.vue'
const { loadTemplates, gridOptions, storeTemplate, deleteTemplate, updateTemplate } = useTemplate();
const { loadTemplateForm, deleteTemplateForm } = useTemplateForm();
const xGrid = ref()
import useRulesStore from '@/store/useRulesStore.ts'
const rulesStore = useRulesStore();

// 模板 CRUD
const templateRef = ref()
const templateState = ref({ template_name: "" })
const templateUpdateState = ref({ id: null, template_name: "" })
const editOpen = ref(false)
const templateUptRef = ref()

const handleNewTemplate = () => {
   templateUpdateState.value = { id: null, template_name: "" }
   editOpen.value = true
}

const editTemplate = (row: any) => {
   templateUpdateState.value = { id: row.id, template_name: row.template_name }
   editOpen.value = true
}

const submitUpdateTemplate = async () => {
   try {
      templateUptRef.value.clearValidate()
      if (templateUpdateState.value.id) {
         await updateTemplate(templateUpdateState.value)
      } else {
         await storeTemplate(templateUpdateState.value)
      }
      templateUptRef.value.resetFormField()
   } catch (error) {
      templateUptRef.value.validate()
   }
   xGrid.value?.commitProxy("query")
   formpartRef.value?.loadTemplateOpts()
   editOpen.value = false
}

// 控件管理
const selectedTemplateId = ref<number | null>(null)
const selectedTemplateName = ref('')
const fields = ref<any[]>([])
const cid = ref<number | null>(null)
const formpartRef = ref()
const open = ref(false)

const fieldTypes: Record<string, string> = {
   text: '文本框', number: '数字', textarea: '文本域', select: '下拉框',
   radio: '单选', checkbox: '复选框', date: '日期', file: '文件',
}
const fieldLabel = (type: string) => fieldTypes[type] || type

const loadTmplForm = async (row: any) => {
   selectedTemplateId.value = row.id
   selectedTemplateName.value = row.template_name
   const { data } = await loadTemplateForm(row.id)
   fields.value = data
}

const editField = (record: any) => {
   cid.value = record.id
   open.value = true
}

const delRecord = async (row: any) => {
   await deleteTemplateForm(row.id)
   if (selectedTemplateId.value) {
      const { data } = await loadTemplateForm(selectedTemplateId.value)
      fields.value = data
   }
}

const handleAddField = () => {
   cid.value = null
   open.value = true
}

// Grid events
const gridEvent: VxeGridListeners<RowVO> = {
   proxyQuery() {
      const grid = xGrid.value
      grid.getTableData().fullData
   },
   proxyDelete() {},
   proxySave() {}
}
</script>

<style scoped>
:deep(.ant-card-body) {
   padding: 12px;
}
</style>
