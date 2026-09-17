<template>
  <div class="panel appointment-panel">
    <div class="panel-head">
      <h2>交换预约 · 确认与撤回闭环</h2>
      <div class="panel-tools">
        <el-select
          :model-value="store.currentActor"
          size="small"
          style="width: 118px"
          @update:model-value="store.setCurrentActor"
        >
          <el-option v-for="name in DEMO_ACTORS" :key="name" :label="`以 ${name}`" :value="name" />
        </el-select>
        <el-button size="small" :loading="store.loading" @click="store.load()">刷新回读</el-button>
        <el-button size="small" type="primary" @click="createVisible = true">发起预约</el-button>
      </div>
    </div>

    <el-alert
      :title="store.lastError"
      type="error"
      show-icon
      :closable="false"
      v-if="store.lastError"
      class="banner"
    />
    <el-alert
      :title="store.lastNotice"
      type="success"
      show-icon
      :closable="false"
      v-if="store.lastNotice"
      class="banner"
    />

    <el-skeleton :loading="store.loading && store.appointments.length === 0" :rows="4" animated>
      <el-empty v-if="store.appointments.length === 0" description="暂无预约" />
      <AppointmentCard
        v-for="appointment in store.appointments"
        v-else
        :key="appointment.id"
        :appointment="appointment"
        :current-actor="store.currentActor"
        :acting="store.actingId === appointment.id"
        @confirm="store.confirm(appointment.id)"
        @withdraw="store.withdraw(appointment.id)"
        @request-cancel="openCancel(appointment.id)"
        @decide="(decision) => store.decideCancel(appointment.id, decision)"
      />
    </el-skeleton>

    <CancelReasonDialog
      :visible="cancelVisible"
      :submitting="store.actingId === cancelTargetId"
      @cancel="cancelVisible = false"
      @submit="submitCancel"
    />
    <CreateAppointmentDialog
      :visible="createVisible"
      :initiator="store.currentActor"
      :submitting="store.actingId === -1"
      @cancel="createVisible = false"
      @submit="submitCreate"
    />
  </div>
</template>

<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { ElMessage } from 'element-plus';
import { DEMO_ACTORS } from '../../constants/appointment.constants';
import { useAppointmentStore } from '../../stores/appointment.store';
import type { CreateAppointmentPayload } from '../../types/domain';
import AppointmentCard from './AppointmentCard.vue';
import CancelReasonDialog from './CancelReasonDialog.vue';
import CreateAppointmentDialog from './CreateAppointmentDialog.vue';

const store = useAppointmentStore();

const cancelVisible = ref(false);
const cancelTargetId = ref<number>(-1);
const createVisible = ref(false);

onMounted(() => {
  void store.load();
});

function openCancel(id: number) {
  cancelTargetId.value = id;
  cancelVisible.value = true;
}

async function submitCancel(reason: string) {
  const ok = await store.requestCancel(cancelTargetId.value, reason);
  if (ok) {
    cancelVisible.value = false;
    ElMessage.success('取消申请已提交，等待对方处理');
  }
}

async function submitCreate(payload: CreateAppointmentPayload) {
  const ok = await store.create(payload);
  if (ok) {
    createVisible.value = false;
    ElMessage.success('预约已发起，双方时间档已占用，等待确认');
  }
}
</script>

<style scoped>
.appointment-panel {
  grid-column: 1 / -1;
}
.panel-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}
.panel-tools {
  display: flex;
  gap: 8px;
  align-items: center;
}
.banner {
  margin: 10px 0;
}
</style>
