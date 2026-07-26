<template>
  <div class="p-2">
    <p>
      <a-button type="primary" @click="addEmp">新建</a-button>
    </p>
    <vxe-grid ref='xGrid' v-bind="{ ...gridOptions, size: 'small' }" v-on="gridEvent">
      <template #action="{ row }">
        <a-space size="small">
          <a-button type="primary" size="small">删除</a-button>
          <a-button type="primary" size="small">编辑</a-button>
        </a-space>
      </template>
      <template #dept="{ row }">
        <span>{{ row.Dept.id == 0 ? "未分配" : row.Dept.dept_name }}</span>
      </template>
    </vxe-grid>
  </div>
</template>

<script setup lang="ts">
const {gridOptions} = useUser()
const router = useRouter()
const rulesStore = useRulesStore()

// TABLE
const xGrid = ref()
const gridEvent: VxeGridListeners<RowVO> = {
  proxyQuery() {
    const grid = xGrid.value
    const data = grid.getTableData().fullData
  },
  proxyDelete() {
  },
  proxySave() {
  }
}

const addEmp = () => {
  router.push({path: "/admin/manage/groupcourse/crt"})
}
const formRef = ref()

const onSubmit = async () => {
  try {
    formRef.value.clearValidate()
  } catch (error) {
    // handled
  }
}

const formModalRef = ref()

</script>

<style scoped></style>
