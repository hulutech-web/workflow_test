<template>
    <div class="comment-thread">
        <div v-for="comment in comments" :key="comment.id" class="comment-item" :class="{ 'ml-8 mt-2': comment.parent_id > 0 }">
            <div class="flex gap-3">
                <!-- Avatar placeholder -->
                <a-avatar size="small" :style="{ backgroundColor: avatarColor(comment.emp_name) }">
                    {{ (comment.emp_name || '?')[0] }}
                </a-avatar>
                <div class="flex-1 min-w-0">
                    <!-- Header -->
                    <div class="flex items-center gap-2 flex-wrap">
                        <span class="font-semibold text-sm">{{ comment.emp_name || '未知用户' }}</span>
                        <a-tag v-if="comment.reply_to_emp_name" color="blue" size="small" class="!text-xs">
                            回复 @{{ comment.reply_to_emp_name }}
                        </a-tag>
                        <span class="text-xs text-gray-400">{{ comment.created_at }}</span>
                    </div>
                    <!-- Content -->
                    <div class="text-sm text-gray-700 mt-1 whitespace-pre-wrap">{{ comment.content }}</div>
                    <!-- Actions -->
                    <div class="mt-1">
                        <a-button
                            v-if="!isReadonly"
                            type="link"
                            size="small"
                            class="!p-0 !h-auto !text-xs"
                            @click="toggleReply(comment)"
                        >
                            {{ replyTarget?.id === comment.id ? '取消回复' : '回复' }}
                        </a-button>
                    </div>
                    <!-- Inline reply editor -->
                    <div v-if="replyTarget?.id === comment.id" class="mt-2">
                        <a-textarea
                            v-model:value="replyContent"
                            :rows="2"
                            size="small"
                            :placeholder="`回复 ${comment.emp_name}...`"
                        />
                        <div class="mt-1">
                            <a-space size="small">
                                <a-button type="primary" size="small" :loading="sending" @click="sendReply(comment)">
                                    发送
                                </a-button>
                                <a-button size="small" @click="cancelReply">取消</a-button>
                            </a-space>
                        </div>
                    </div>
                </div>
            </div>
            <!-- Recursive children -->
            <CommentThread
                v-if="comment.children?.length"
                :comments="comment.children"
                :entry-id="entryId"
                :proc-id="procId"
                :is-readonly="isReadonly"
                @comment-added="$emit('commentAdded')"
            />
        </div>
    </div>
</template>

<script setup lang="ts">
import { message } from 'ant-design-vue'
const { addComment } = useProc()

const props = withDefaults(defineProps<{
    comments: any[]
    entryId?: string | number
    procId?: string | number
    isReadonly?: boolean
}>(), {
    isReadonly: false
})

defineEmits<{ commentAdded: [] }>()

const replyTarget = ref<any>(null)
const replyContent = ref('')
const sending = ref(false)

const toggleReply = (comment: any) => {
    if (replyTarget.value?.id === comment.id) {
        replyTarget.value = null
        replyContent.value = ''
    } else {
        replyTarget.value = comment
        replyContent.value = ''
    }
}

const cancelReply = () => {
    replyTarget.value = null
    replyContent.value = ''
}

const sendReply = async (parentComment: any) => {
    if (!replyContent.value.trim()) {
        message.warning('请输入回复内容')
        return
    }
    sending.value = true
    try {
        await addComment({
            entry_id: props.entryId,
            proc_id: props.procId || 0,
            content: replyContent.value,
            parent_id: parentComment.id,
            reply_to_emp_id: parentComment.emp_id,
            reply_to_emp_name: parentComment.emp_name,
        })
        replyContent.value = ''
        replyTarget.value = null
        message.success('回复成功')
    } catch (e) {
        // error handled by interceptor
    } finally {
        sending.value = false
    }
}

const avatarColors = ['#1890ff', '#52c41a', '#fa8c16', '#722ed1', '#eb2f96', '#13c2c2', '#f5222d', '#2f54eb']
const avatarColor = (name: string) => {
    if (!name) return avatarColors[0]
    let hash = 0
    for (let i = 0; i < name.length; i++) {
        hash = name.charCodeAt(i) + ((hash << 5) - hash)
    }
    return avatarColors[Math.abs(hash) % avatarColors.length]
}
</script>

<script lang="ts">
export default {
    name: 'CommentThread',
}
</script>
