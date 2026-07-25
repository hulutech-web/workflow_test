<template>
    <div class="p-4">
        <div class="mb-4">
            <a-button type="link" @click="goBack">
                <ArrowLeftOutlined /> 返回
            </a-button>
        </div>
        <div class="flex justify-center">
            <div class="w-full" style="max-width: 700px">
                <a-card>
                    <template #title>
                        <div class="text-center">
                            <span class="text-lg font-semibold">流程：{{ flow.flow_name }}</span>
                            <a-tag v-if="flow.Template" color="blue" class="ml-2">
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
        console.log(values)
        await storeEntry(values)
    } catch (error) {
        huluFormRef.value.validate()
    }
}
init();

</script>

<style></style>
