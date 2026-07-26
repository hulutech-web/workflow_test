<template>
  <div class="p-2">
    <p>归档审批</p>

    <vxe-grid ref="xGrid" v-bind="gridOptions">
      <template #title_default="{ row }">
        <span>{{ row.title || `归档${row.id}` }}</span>
      </template>
      <template #status="{ row }">
        <a-badge v-if="row.status == 9" status="success" text="已完成"/>
        <a-badge v-if="row.status == -1" status="error" text="已驳回"/>
        <a-badge v-if="row.status == -2" status="error" text="已撤回"/>
      </template>
      <template #action="{ row }">
        <a-button type="primary" ghost size="small" @click="viewDetail(row)">详情</a-button>
      </template>
    </vxe-grid>
  </div>
</template>

<script setup lang="ts">
const { gridOptions } = useEntryArchive()
const router = useRouter()

const xGrid = ref()

const viewDetail = (row: any) => {
  router.push({ path: `/admin/base/flow/archive/${row.id}` })
}

onMounted(() => {
  xGrid.value?.commitProxy('query')
})
</script>
