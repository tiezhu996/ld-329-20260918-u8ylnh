<template>
  <div class="appointment-actions">
    <template v-if="isParty">
      <template v-if="appointment.state === 'pending'">
        <el-button v-if="!hasConfirmed" type="primary" size="small" @click="emit('confirm')">确认预约</el-button>
        <el-button v-if="isInitiator" type="danger" size="small" plain @click="emit('withdraw')">撤回预约</el-button>
        <span v-if="isInitiator" class="muted">待对方确认，仅发起人可撤回</span>
        <span v-else-if="hasConfirmed" class="muted">已确认，等待对方确认</span>
      </template>

      <el-button v-else-if="appointment.state === 'confirmed'" type="warning" size="small" plain @click="emit('request-cancel')">
        申请取消
      </el-button>

      <template v-else-if="appointment.state === 'cancel_pending' && appointment.cancelRequest">
        <template v-if="appointment.cancelRequest.by === currentUser">
          <span class="muted">取消申请待对方处理：{{ appointment.cancelRequest.reason }}</span>
        </template>
        <template v-else>
          <span class="muted">对方申请取消：{{ appointment.cancelRequest.reason }}</span>
          <el-button type="danger" size="small" @click="emit('respond-cancel', 'accept')">同意取消</el-button>
          <el-button size="small" @click="emit('respond-cancel', 'reject')">拒绝取消</el-button>
        </template>
      </template>
    </template>
    <span v-else class="muted">仅预约双方可操作</span>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import type { Appointment, CancelAction } from '../types/appointment';

const props = defineProps<{ appointment: Appointment; currentUser: string }>();

const emit = defineEmits<{
  confirm: [];
  withdraw: [];
  'request-cancel': [];
  'respond-cancel': [action: CancelAction];
}>();

const isInitiator = computed(() => props.appointment.initiator === props.currentUser);
const isParty = computed(
  () => isInitiator.value || props.appointment.participant === props.currentUser,
);
const hasConfirmed = computed(() => props.appointment.confirmations.includes(props.currentUser));
</script>

<style scoped>
.appointment-actions {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 8px;
}
</style>
