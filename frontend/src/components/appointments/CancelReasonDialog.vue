<template>
  <el-dialog
    :model-value="visible"
    title="提交取消申请"
    width="420px"
    :close-on-click-modal="false"
    @update:model-value="(value: boolean) => !value && emitCancel"
    @close="emitCancel"
  >
    <el-alert
      type="warning"
      :closable="false"
      show-icon
      title="预约已双方确认，不能直接撤回。提交取消原因后需等待对方处理，对方同意前时间档继续占用。"
      class="cancel-tip"
    />
    <el-input
      v-model="reason"
      type="textarea"
      :rows="3"
      maxlength="300"
      show-word-limit
      placeholder="请说明取消原因（至少 4 个字），例如：临时答辩冲突，需要改期"
      @input="errorMessage = ''"
    />
    <p v-if="errorMessage" class="cancel-error">{{ errorMessage }}</p>
    <template #footer>
      <el-button @click="emitCancel">再想想</el-button>
      <el-button type="danger" :loading="submitting" @click="submit">
        提交取消申请
      </el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import {
  MAX_CANCEL_REASON_LENGTH,
  MIN_CANCEL_REASON_LENGTH,
} from '../../constants/appointment.constants';

const props = defineProps<{
  visible: boolean;
  submitting: boolean;
}>();

const emit = defineEmits<{
  (e: 'cancel'): void;
  (e: 'submit', reason: string): void;
}>();

const reason = ref('');
const errorMessage = ref('');

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      reason.value = '';
      errorMessage.value = '';
    }
  },
);

function emitCancel() {
  emit('cancel');
}

function submit() {
  const trimmed = reason.value.trim();
  if (trimmed.length < MIN_CANCEL_REASON_LENGTH) {
    errorMessage.value = `取消原因至少 ${MIN_CANCEL_REASON_LENGTH} 个字`;
    return;
  }
  if (trimmed.length > MAX_CANCEL_REASON_LENGTH) {
    errorMessage.value = `取消原因不能超过 ${MAX_CANCEL_REASON_LENGTH} 个字`;
    return;
  }
  emit('submit', trimmed);
}
</script>

<style scoped>
.cancel-tip {
  margin-bottom: 12px;
}
.cancel-error {
  margin: 6px 0 0;
  color: var(--el-color-danger);
  font-size: 12px;
}
</style>
