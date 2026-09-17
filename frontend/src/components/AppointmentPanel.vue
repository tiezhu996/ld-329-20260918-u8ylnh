<template>
  <div class="panel">
    <div class="panel-title">
      <h2>{{ APPOINTMENT_PANEL_TITLE }}</h2>
      <el-button size="small" text :loading="loading" @click="loadAppointments">刷新</el-button>
    </div>
    <el-timeline v-if="appointments.length">
      <AppointmentItem
        v-for="item in appointments"
        :key="item.id"
        :appointment="item"
        :current-user="currentUser"
        @confirm="run(() => confirmAppointment(item.id, currentUser), '已确认预约')"
        @withdraw="onWithdraw(item)"
        @request-cancel="openCancelDialog(item)"
        @respond-cancel="(action) => onRespondCancel(item, action)"
      />
    </el-timeline>
    <el-empty v-else description="暂无预约" :image-size="60" />

    <el-dialog v-model="cancelDialogVisible" :title="CANCEL_DIALOG_TITLE" width="420px">
      <el-input
        v-model="cancelReason"
        type="textarea"
        :rows="3"
        :placeholder="CANCEL_REASON_PLACEHOLDER"
      />
      <template #footer>
        <el-button @click="cancelDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitCancel">提交取消申请</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage, ElMessageBox } from 'element-plus';
import AppointmentItem from './AppointmentItem.vue';
import {
  confirmAppointment,
  fetchAppointments,
  requestCancel,
  respondCancel,
  withdrawAppointment,
} from '../services/appointment.service';
import {
  APPOINTMENT_PANEL_TITLE,
  CANCEL_DIALOG_TITLE,
  CANCEL_REASON_PLACEHOLDER,
  CANCEL_REASON_REQUIRED_TEXT,
  WITHDRAW_CONFIRM_TEXT,
} from '../constants/appointment.constants';
import { AppException } from '../errors/AppException';
import { logger } from '../logger/logger';
import type { Appointment, CancelAction } from '../types/appointment';

const props = defineProps<{ currentUser: string }>();
const emit = defineEmits<{ changed: [] }>();

const appointments = ref<Appointment[]>([]);
const loading = ref(false);
const submitting = ref(false);
const cancelDialogVisible = ref(false);
const cancelReason = ref('');
const cancelTarget = ref<Appointment | null>(null);

async function loadAppointments() {
  loading.value = true;
  try {
    appointments.value = await fetchAppointments();
  } catch (err) {
    showError(err);
  } finally {
    loading.value = false;
  }
}

// 每个操作完成后重新回读列表，保证页面与后端终态一致。
async function run(action: () => Promise<Appointment>, successText: string) {
  submitting.value = true;
  try {
    await action();
    ElMessage.success(successText);
  } catch (err) {
    showError(err);
  } finally {
    submitting.value = false;
    await loadAppointments();
    emit('changed');
  }
}

async function onWithdraw(item: Appointment) {
  try {
    await ElMessageBox.confirm(WITHDRAW_CONFIRM_TEXT, '撤回预约', { type: 'warning' });
  } catch {
    return;
  }
  await run(() => withdrawAppointment(item.id, props.currentUser), '已撤回预约，双方时间档已释放');
}

function openCancelDialog(item: Appointment) {
  cancelTarget.value = item;
  cancelReason.value = '';
  cancelDialogVisible.value = true;
}

async function submitCancel() {
  if (!cancelReason.value.trim()) {
    ElMessage.warning(CANCEL_REASON_REQUIRED_TEXT);
    return;
  }
  const target = cancelTarget.value;
  if (!target) {
    return;
  }
  cancelDialogVisible.value = false;
  await run(
    () => requestCancel(target.id, props.currentUser, cancelReason.value.trim()),
    '已提交取消申请，等待对方处理',
  );
}

async function onRespondCancel(item: Appointment, action: CancelAction) {
  const text = action === 'accept' ? '已同意取消，双方时间档已释放' : '已拒绝取消，预约恢复生效';
  await run(() => respondCancel(item.id, props.currentUser, action), text);
}

function showError(err: unknown) {
  const message = err instanceof AppException ? err.message : '操作失败，请稍后重试';
  logger.warn('appointment action failed', err);
  ElMessage.error(message);
}

onMounted(loadAppointments);
</script>

<style scoped>
.panel-title {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 14px;
}
.panel-title h2 {
  margin: 0;
  font-size: 20px;
}
</style>
