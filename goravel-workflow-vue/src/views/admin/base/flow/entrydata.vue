<template>
    <div>
        <a-row>
            <a-col :span="8"></a-col>
            <a-col :span="8">
                <div class="p-3">
                    <a-watermark :content="entry.status === -1 || entry.status === -2 ? '已驳回' : '已发起'">
                        <div>
                            <a-card>
                                <p class="text-2xl font-bold mb-3 text-center">
                                    <span>{{ flow.flow_name }}</span>
                                <div>
                                    <a-tag v-if="flow.Template" color="blue" class="ml-2">
                                        {{ flow.Template.template_name }}
                                    </a-tag>
                                </div>
                                </p>

                                <a-alert v-if="canEdit" message="该流程已被驳回，修改后重新提交即可" type="warning" show-icon class="mb-3" />

                                <Form :fields="fillFields" @submit="onSubmit" :entryDatas="entryDatas"
                                    ref="huluFormRef">
                                    <template v-if="!canEdit">
                                        <div></div>
                                    </template>
                                </Form>
                            </a-card>
                        </div>
                    </a-watermark>

                </div>
            </a-col>
            <a-col :span="8"></a-col>

        </a-row>
    </div>

</template>

<script setup lang='ts'>
const { loadFlowEntryConfig, storeEntry, updateEntry, getEntryData } = useEntry();
const route = useRoute();
const flow_id = route.params.flow_id
const entry_id = route.params.entry_id;
const fillFields = ref([]);
const flow = ref({})
const entry = ref({})
const canEdit = ref(false)

const init = async () => {
    if (flow_id) {
        const { data } = await loadFlowEntryConfig(flow_id);
        flow.value = data
        fillFields.value = data.Template.TemplateForms
    }
    if (entry_id) {
        await loadEntryDatas()
    }
}
const entryDatas = ref([])
const loadEntryDatas = async () => {
    const { data } = await getEntryData(entry_id)
    entryDatas.value = data.entrydata
    entry.value = data.entry
    canEdit.value = data.entry.status === -1 || data.entry.status === -2
}

const huluFormRef = ref()
const onSubmit = async (values) => {
    if (canEdit.value) {
        try {
            huluFormRef.value.clearValidate()
            values.title = entry.value.title
            await updateEntry(entry_id, values)
        } catch (error) {
            huluFormRef.value.validate()
        }
    } else {
        values.flow_id = +flow_id
        try {
            huluFormRef.value.clearValidate()
            await storeEntry(values)
        } catch (error) {
            huluFormRef.value.validate()
        }
    }
}
init();

</script>

<style></style>