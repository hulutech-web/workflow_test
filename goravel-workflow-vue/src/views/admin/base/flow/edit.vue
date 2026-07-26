<template>
    <div class="p-2">
        <div class="mb-2">
            <a-button type="link" size="small" @click="goBack">
                <ArrowLeftOutlined /> 返回
            </a-button>
        </div>
        <div class="flex justify-center">
            <div style="max-width: 700px; width: 100%">
                <a-card size="small" title="编辑流程">
                    <Form :id="id"/>
                </a-card>
            </div>
        </div>
    </div>
</template>

<script setup lang='ts'>
const route = useRoute()
let id = route.params.id
import Form from './form.vue'
const router = useRouter()
const goBack = () => router.back()
</script>

<style></style>
