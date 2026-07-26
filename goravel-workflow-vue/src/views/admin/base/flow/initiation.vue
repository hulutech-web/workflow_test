<template>
    <div class="p-2">
        <div class="mb-2">
            <a-button type="link" size="small" @click="goBack">
                <ArrowLeftOutlined /> 返回
            </a-button>
        </div>
        <div class="flex justify-center">
            <div style="max-width: 700px; width: 100%">
                <a-card size="small">
                    <template #title>
                        <div class="text-center">
                            <span class="text-base font-semibold">流程：{{ flow.flow_name }}</span>
                            <a-tag v-if="flow.Template" color="blue" class="ml-1">
                                {{ flow.Template.template_name }}
                            </a-tag>
                        </div>
                    </template>

                    <Form :fields="fillFields" @submit="onSubmit" ref="huluFormRef"></Form>
                </a-card>
            </div>
        </div>
    </div>
</template>

<script setup lang='ts'>
const { loadFlowEntryConfig, storeEntry } = useEntry();
const route = useRoute();
const id = route.params.id;
const fillFields = ref([]);
const flow = ref({})
const router = useRouter()
const goBack = () => router.back()

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
        await storeEntry(values)
    } catch (error) {
        huluFormRef.value.validate()
    }
}
init();

</script>

<style></style>
