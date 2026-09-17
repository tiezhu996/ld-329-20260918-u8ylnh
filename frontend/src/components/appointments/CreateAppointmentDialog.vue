<template>
  <el-dialog
    :model-value="visible"
    title="发起交换预约"
    width="460px"
    :close-on-click-modal="false"
    @update:model-value="(value: boolean) => !value && emit('cancel')"
    @close="emit('cancel')"
  >
    <el-form label-position="top">
      <el-form-item label="发起人">
        <el-input :model-value="initiator" disabled />
      </el-form-item>
      <el-form-item label="响应人" required>
        <el-input v-model="form.responder" placeholder="对方姓名" maxlength="80" />
      </el-form-item>
      <el-form-item label="交换时间" required>
        <el-input v-model="form.time" placeholder="如：周六 10:00" maxlength="80" />
      </el-form-item>
      <el-form-item label="交换地点" required>
        <el-select v-model="form.place" placeholder="线上 / 线下" style="width: 100%">
          <el-option label="线上会议室" value="线上会议室" />
          <el-option label="东校区湖边" value="东校区湖边" />
          <el-option label="西校区琴房" value="西校区琴房" />
          <el-option label="中心校区自习室" value="中心校区自习室" />
        </el-select>
      </el-form-item>
      <el-form-item label="时间档（创建后立即占用双方时间档）" required>
        <el-select
          v-model="form.slots"
          multiple
          filterable
          allow-create
          default-first-option
          placeholder="选择或输入时间档"
          style="width: 100%"
        >
          <el-option v-for="slot in suggestedSlots" :key="slot" :label="slot" :value="slot" />
        </el-select>
      </el-form-item>
      <el-form-item label="协商议程">
        <el-input v-model="form.agenda" type="textarea" :rows="2" maxlength="500" show-word-limit />
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="emit('cancel')">取消</el-button>
      <el-button type="primary" :loading="submitting" @click="submit">发起预约</el-button>
    </template>
  </el-dialog>
</template>

<script setup lang="ts">
import { reactive, watch } from 'vue';
import type { CreateAppointmentPayload } from '../../types/domain';

const suggestedSlots = ['周一晚', '周二晚', '周三晚', '周四晚', '周五晚', '周六上午', '周六下午', '周日全天'];

const props = defineProps<{
  visible: boolean;
  initiator: string;
  submitting: boolean;
}>();

const emit = defineEmits<{
  (e: 'cancel'): void;
  (e: 'submit', payload: CreateAppointmentPayload): void;
}>();

const form = reactive({
  responder: '',
  time: '',
  place: '',
  slots: [] as string[],
  agenda: '',
});

watch(
  () => props.visible,
  (visible) => {
    if (visible) {
      form.responder = '';
      form.time = '';
      form.place = '';
      form.slots = [];
      form.agenda = '';
    }
  },
);

function submit() {
  emit('submit', {
    initiator: props.initiator,
    responder: form.responder.trim(),
    time: form.time.trim(),
    place: form.place.trim(),
    agenda: form.agenda.trim(),
    slots: form.slots.map((slot) => slot.trim()).filter(Boolean),
  });
}
</script>
