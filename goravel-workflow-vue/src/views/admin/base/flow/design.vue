<template>
    <div class="p-2">

        <div class="mb-2">
          <a-page-header
              style="border: 1px solid rgb(235, 237, 240)"
              title="流程设计"
              sub-title="自由拖拽与连线，配置位置与现实图标和配色"
              @back="goBack"
          >
            <template #extra>
              <a-button type="primary" size="small" @click="saveDesign">保存位置</a-button>
              <a-button type="primary" size="small" @click="publishDesign">发布流程</a-button>
            </template>
          </a-page-header>
        </div>
            <div id="flow-chart-container">
                <hulu-menu :flow_id="(+id)" :init="initAll" ref="menuRef"/>

                <div v-for="(node, nodeId) in nodeList" :key="node.id"
                     :class="'node' + (node.process_to ? ' source-node' : '')" :id="'node-' + node.id" :style="node.style">
                  <div class="flex justify-center align-items-center node-element" :id="`menu-${node.id}`">
                    <HuluIcon :id="`node-line-${node.id}-pointer`" fontSize="20px" :name="node.icon" color="#66CDAA"/>
                    <span class="font-bold text-sm">{{ node.process_name }}</span>
                    <a-button type="primary" style="color:#ffffff;z-index:20;background-color: #FFA500;" @click="setProcess(node)"
                              shape="circle" size="small">
                      <FormOutlined class="node-setting"/>
                    </a-button>
                  </div>
                </div>
            </div>

        <a-modal v-model:open="open" style="position: relative;" width="1200px" :footer="false" title="节点设计">
          <attrform :attrs="attrs" @updProcess="updProcess"/>
        </a-modal>
    </div>
</template>

<script setup lang='ts'>
import initFlowChart from './flow'

const route = useRoute()
const router = useRouter();
const {loadFlowDesign, publishFlow} = useFlow()
const {updateFlowlink} = useFlowlink();
const {loadAttributes, updateProcess} = useProcess()
const id = route.params.id
const jsplumbJSON = ref({})
const nodeList = ref([])
const flow = ref({})
const menuRef = ref({})
const open = ref(false)

const goBack = () => router.back()

const init = async () => {
  const {data} = await loadFlowDesign(+id)
  flow.value = data
  if (data.jsplumb) {
    jsplumbJSON.value = JSON.parse(data.jsplumb)
    nodeList.value = jsplumbJSON.value.list
    Object.entries(nodeList.value).map(([key, value]) => {
      value.flow_id = +id
    })
  }
}


onMounted(async () => {
  await initAll()
})

const updProcess = async (val) => {
  await updateProcess(process_id.value, val)
  // 保存后重新加载流程数据，确保 flow.value.jsplumb 是最新的
  await initAll()
  open.value = false
}

const initAll = async () => {
  await init()
  await initFlowChart(jsplumbJSON.value, getNewestNodes)
}
const saveDesign = async () => {
  console.log(JSON.parse(flow.value.jsplumb))
  await updateFlowlink(flow.value)
}

const attrs = ref({})
const process_id = ref(0)
const setProcess = async (node) => {
  open.value = true
  const {data} = await loadAttributes(node.id)
  process_id.value = node.id
  attrs.value = data
}
const getNewestNodes = async (nodes) => {
  let newJsplumb = {
    total: nodes.length,
    list: ""
  }
  let list = Object.create(null)
  for (let i = 0; i < newJsplumb.total; i++) {
    let node = nodes[i]
    list[node.id + ""] = node
  }
  newJsplumb.list = list
  flow.value.jsplumb = JSON.stringify(newJsplumb)
}

const publishDesign = async () => {
  await publishFlow({flow_id: flow.value.id})
}
</script>


<style scoped>
#flow-chart-container {
  width: 100%;
  height: 500px;
  border: 1px solid #ccc;
  position: relative;
  background-image: linear-gradient(90deg, rgba(200, 200, 200, 0.2) 1px, transparent 1px),
  linear-gradient(180deg, rgba(200, 200, 200, 0.2) 1px, transparent 1px);
  background-size: 12px 12px;
}

.node {
  position: absolute;
  text-align: center;
}

.node-element {
  background-color: #FFFACD;
  border: 1px solid #FFFACD;
  border-radius: 4px;
  padding: 4px;
  font-size: 12px;
}

.node-element:hover {
  background-color: #08140e;
  cursor: move;
}

.node-setting {
  cursor: pointer;
}
</style>
