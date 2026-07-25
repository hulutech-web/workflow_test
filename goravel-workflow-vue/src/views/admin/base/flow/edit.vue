<template>
    <div class="p-4">
        <div class="mb-4">
            <a-button type="link" @click="goBack">
                <ArrowLeftOutlined /> 返回
            </a-button>
        </div>
        <div class="flex justify-center">
            <div class="w-full" style="max-width: 700px">
                <a-card title="编辑流程">
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