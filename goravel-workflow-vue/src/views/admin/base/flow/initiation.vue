<template>
    <div>
        <a-row>
            <a-col :span="8"></a-col>
            <a-col :span="8">
                <div class="p-3">
                    <a-card>
                        <p class="text-xl font-bold mb-3 text-center">
                            <span>流程：{{ flow.flow_name }}</span>
                        <div>
                            <a-tag v-if="flow.Template" color="blue" class="ml-2">
                                {{ flow.Template.template_name }}
                            </a-tag>
                        </div>
                        </p>

                        <Form :fields="fillFields" @submit="onSubmit" ref="huluFormRef"></Form>
                    </a-card>
                </div>
            </a-col>
            <a-col :span="8"></a-col>

        </a-row>
    </div>

</template>

<script setup lang='ts'>
const { loadFlowEntryConfig, storeEntry } = useEntry();
const route = useRoute();
const id = route.params.id;
const fillFields = ref([]);
const flow = ref({})
const init = async () => {
    if (id) {
        const { data } = await loadFlowEntryConfig(id);
        flow.value = data
        fillFields.value = data.Template.TemplateForms
    }
}

const huluFormRef = ref()
const onSubmit = async (values) => {
    values.flow_id = +id
    try {
        huluFormRef.value.clearValidate()
        console.log(values)
        await storeEntry(values)
    } catch (error) {
        huluFormRef.value.validate()
    }
}
init();

</script>

<style></style>