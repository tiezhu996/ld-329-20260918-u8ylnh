<template>
  <div class="appt-card" :class="{ 'is-terminal': terminal }">
    <div class="appt-head">
      <div>
        <strong>#{{ appointment.id }} {{ appointment.pair }}</strong>
        <el-tag :type="tagType" size="small" effect="dark" class="status-tag">
          {{ appointment.statusText }}
        </el-tag>
      </div>
      <el-tag size="small" :type="appointment.slotsHeld ? 'warning' : 'info'" effect="plain">
        {{ appointment.slotsHeld ? '时间档占用中' : '时间档已释放' }}
      </el-tag>
    </div>

    <p class="appt-meta">
      {{ appointment.time }} · {{ appointment.place }}
      <span v-for="slot in appointment.slots" :key="slot" class="slot-chip">{{ slot }}</span>
    </p>
    <p v-if="appointment.agenda" class="muted">议程：{{ appointment.agenda }}</p>

    <div class="confirm-row">
      <el-tag
        size="small"
        :type="appointment.initiatorConfirmed ? 'success' : 'info'"
        effect="plain"
      >
        发起人 {{ appointment.initiator }}：{{ appointment.initiatorConfirmed ? '已确认' : '待确认' }}
      </el-tag>
      <el-tag
        size="small"
        :type="appointment.responderConfirmed ? 'success' : 'info'"
        effect="plain"
      >
        响应人 {{ appointment.responder }}：{{ appointment.responderConfirmed ? '已确认' : '待确认' }}
      </el-tag>
    </div>

    <el-alert
      v-if="appointment.status === APPOINTMENT_STATUS.CANCEL_REQUESTED && appointment.cancel"
      type="warning"
      :closable="false"
      show-icon
      class="cancel-banner"
      :title="`${appointment.cancel.by} 申请取消：${appointment.cancel.reason}`"
      description="取消待对方处理；对方同意后时间档才会释放。"
    />

    <div v-if="participant && !terminal" class="action-row">
      <template v-if="canConfirm">
        <el-button
          type="primary"
          size="small"
          :loading="acting"
          @click="emit('confirm')"
        >
          我已确认
        </el-button>
      </template>

      <template v-if="canWithdraw">
        <el-popconfirm
          title="撤回后双方时间档立即释放，且不能恢复。确认撤回？"
          confirm-button-text="撤回"
          cancel-button-text="再想想"
          confirm-button-type="danger"
          @confirm="emit('withdraw')"
        >
          <template #reference>
            <el-button type="danger" plain size="small" :loading="acting">
              撤回预约
            </el-button>
          </template>
        </el-popconfirm>
        <span class="rule-hint">待确认阶段仅发起人可撤回</span>
      </template>

      <template v-if="canRequestCancel">
        <el-button type="danger" plain size="small" :loading="acting" @click="emit('request-cancel')">
          申请取消
        </el-button>
        <span class="rule-hint">已确认锁定，只能提交取消原因等待对方处理</span>
      </template>

      <template v-if="canDecideCancel">
        <el-button type="danger" size="small" :loading="acting" @click="emit('decide', 'approve')">
          同意取消
        </el-button>
        <el-button size="small" :loading="acting" @click="emit('decide', 'reject')">
          拒绝取消
        </el-button>
      </template>
    </div>

    <p v-else-if="!participant" class="rule-hint">以预约双方身份登录后可操作</p>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue';
import {
  APPOINTMENT_STATUS,
  STATUS_TAG_TYPE,
} from '../../constants/appointment.constants';
import type { CancelDecision } from '../../constants/appointment.constants';
import type { Appointment } from '../../types/domain';

const props = defineProps<{
  appointment: Appointment;
  currentActor: string;
  acting: boolean;
}>();

const emit = defineEmits<{
  (e: 'confirm'): void;
  (e: 'withdraw'): void;
  (e: 'request-cancel'): void;
  (e: 'decide', decision: CancelDecision): void;
}>();

const participant = computed(
  () =>
    props.appointment.initiator === props.currentActor ||
    props.appointment.responder === props.currentActor,
);
const terminal = computed(
  () =>
    props.appointment.status === APPOINTMENT_STATUS.WITHDRAWN ||
    props.appointment.status === APPOINTMENT_STATUS.CANCELLED,
);
const tagType = computed(() => STATUS_TAG_TYPE[props.appointment.status]);
const actorConfirmed = computed(() => {
  if (props.appointment.initiator === props.currentActor) {
    return props.appointment.initiatorConfirmed;
  }
  if (props.appointment.responder === props.currentActor) {
    return props.appointment.responderConfirmed;
  }
  return false;
});
const awaitingConfirm = computed(
  () =>
    props.appointment.status === APPOINTMENT_STATUS.PENDING ||
    props.appointment.status === APPOINTMENT_STATUS.CONFIRMED_PARTIAL,
);
const canConfirm = computed(
  () => participant.value && awaitingConfirm.value && !actorConfirmed.value,
);
const canWithdraw = computed(
  () =>
    participant.value &&
    awaitingConfirm.value &&
    props.appointment.initiator === props.currentActor,
);
const canRequestCancel = computed(
  () =>
    participant.value &&
    props.appointment.status === APPOINTMENT_STATUS.CONFIRMED,
);
const canDecideCancel = computed(
  () =>
    participant.value &&
    props.appointment.status === APPOINTMENT_STATUS.CANCEL_REQUESTED &&
    props.appointment.cancel?.by !== props.currentActor,
);
</script>

<style scoped>
.appt-card {
  border: 1px solid var(--el-border-color-lighter);
  border-radius: 10px;
  padding: 12px 14px;
  margin-bottom: 12px;
  background: var(--el-bg-color);
}
.appt-card.is-terminal {
  opacity: 0.72;
  background: var(--el-fill-color-light);
}
.appt-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}
.status-tag {
  margin-left: 8px;
}
.appt-meta {
  margin: 8px 0 4px;
  font-size: 13px;
}
.slot-chip {
  margin-left: 6px;
  padding: 0 6px;
  border-radius: 6px;
  background: var(--el-fill-color);
  font-size: 12px;
}
.confirm-row {
  display: flex;
  gap: 8px;
  margin: 8px 0;
  flex-wrap: wrap;
}
.action-row {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-top: 10px;
  flex-wrap: wrap;
}
.cancel-banner {
  margin-top: 8px;
}
.rule-hint {
  font-size: 12px;
  color: var(--el-text-color-secondary);
}
.muted {
  color: var(--el-text-color-secondary);
  font-size: 13px;
  margin: 4px 0;
}
</style>
