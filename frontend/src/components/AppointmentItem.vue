<template>
  <el-timeline-item :timestamp="appointment.time" placement="top">
    <div class="appointment-item">
      <div class="appointment-item__head">
        <strong>{{ appointment.pair }}</strong>
        <el-tag :type="tagType" size="small">{{ appointment.status }}</el-tag>
      </div>
      <p>{{ appointment.place }} · 发起人 {{ appointment.initiator }}</p>
      <p class="muted">{{ appointment.agenda }}</p>
      <div class="tag-row">
        <el-tag
          v-for="name in appointment.confirmations"
          :key="name"
          type="success"
          effect="plain"
          size="small"
        >{{ name }} 已确认</el-tag>
        <el-tag v-if="appointment.slotReleased" type="info" effect="plain" size="small">
          {{ SLOT_RELEASED_TEXT }}
        </el-tag>
      </div>
      <AppointmentActions
        :appointment="appointment"
        :current-user="currentUser"
        @confirm="emit('confirm')"
        @withdraw="emit('withdraw')"
        @request-cancel="emit('request-cancel')"
        @respond-cancel="(action) => emit('respond-cancel', action)"
      />
    </div>
  </el-timeline-item>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import AppointmentActions from './AppointmentActions.vue';
import { APPOINTMENT_STATE_TAG, SLOT_RELEASED_TEXT } from '../constants/appointment.constants';
import type { Appointment, CancelAction } from '../types/appointment';

const props = defineProps<{ appointment: Appointment; currentUser: string }>();

const emit = defineEmits<{
  confirm: [];
  withdraw: [];
  'request-cancel': [];
  'respond-cancel': [action: CancelAction];
}>();

const tagType = computed(() => APPOINTMENT_STATE_TAG[props.appointment.state] ?? 'info');
</script>

<style scoped>
.appointment-item__head {
  display: flex;
  align-items: center;
  gap: 10px;
}
.appointment-item p {
  margin: 6px 0;
}
</style>
