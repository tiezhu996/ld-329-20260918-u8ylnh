import { defineStore } from 'pinia';
import { APPOINTMENT_STATUS, DEMO_ACTORS } from '../constants/appointment.constants';
import { logger } from '../logger/logger';
import { appointmentApi } from '../services/appointment.service';
import type {
  Appointment,
  AppointmentActionResult,
  CreateAppointmentPayload,
} from '../types/domain';
import type { CancelDecision } from '../constants/appointment.constants';

// 交换预约确认 / 撤回 / 取消闭环的前端状态。
// 服务端是状态机的唯一权威；前端只负责回读结果，重复请求以服务端终态为准。
interface AppointmentState {
  currentActor: string;
  appointments: Appointment[];
  loading: boolean;
  actingId: number | null;
  lastError: string;
  lastNotice: string;
}

export const useAppointmentStore = defineStore('appointment', {
  state: (): AppointmentState => ({
    currentActor: DEMO_ACTORS[0],
    appointments: [],
    loading: false,
    actingId: null,
    lastError: '',
    lastNotice: '',
  }),

  getters: {
    byId: (state) => (id: number) =>
      state.appointments.find((item) => item.id === id),
  },

  actions: {
    setCurrentActor(actor: string) {
      this.currentActor = actor;
    },

    async load(): Promise<void> {
      this.loading = true;
      this.lastError = '';
      try {
        this.appointments = await appointmentApi.list();
      } catch (err) {
        this.lastError = err instanceof Error ? err.message : '预约加载失败';
        logger.error('load appointments failed', err);
      } finally {
        this.loading = false;
      }
    },

    // 所有操作落同一入口：成功后用服务端返回的快照原地替换，保证刷新语义。
    async runAction(
      id: number,
      action: () => Promise<AppointmentActionResult>,
    ): Promise<boolean> {
      this.actingId = id;
      this.lastError = '';
      this.lastNotice = '';
      try {
        const result = await action();
        this.replace(result.appointment);
        if (result.changed && result.note) {
          this.lastNotice = result.note;
        }
        return true;
      } catch (err) {
        this.lastError = err instanceof Error ? err.message : '操作失败';
        // 失败也回读一次，避免本地展示与终态脱节（例如并发时对方先撤回）。
        await this.load();
        return false;
      } finally {
        this.actingId = null;
      }
    },

    replace(updated: Appointment) {
      const index = this.appointments.findIndex((item) => item.id === updated.id);
      if (index === -1) {
        this.appointments.push(updated);
      } else {
        this.appointments[index] = updated;
      }
    },

    confirm(id: number) {
      return this.runAction(id, () => appointmentApi.confirm(id, this.currentActor));
    },

    withdraw(id: number) {
      return this.runAction(id, () => appointmentApi.withdraw(id, this.currentActor));
    },

    requestCancel(id: number, reason: string) {
      return this.runAction(id, () =>
        appointmentApi.requestCancel(id, this.currentActor, reason),
      );
    },

    decideCancel(id: number, decision: CancelDecision) {
      return this.runAction(id, () =>
        appointmentApi.decideCancel(id, this.currentActor, decision),
      );
    },

    async create(payload: CreateAppointmentPayload): Promise<boolean> {
      return this.runAction(-1, () => appointmentApi.create(payload));
    },

    // 便捷判定，供按钮显隐使用；真正的拦截在服务端。
    isInitiator(appt: Appointment): boolean {
      return appt.initiator === this.currentActor;
    },
    isParticipant(appt: Appointment): boolean {
      return (
        appt.initiator === this.currentActor ||
        appt.responder === this.currentActor
      );
    },
    actorConfirmed(appt: Appointment): boolean {
      if (appt.initiator === this.currentActor) {
        return appt.initiatorConfirmed;
      }
      if (appt.responder === this.currentActor) {
        return appt.responderConfirmed;
      }
      return false;
    },
    isTerminal(appt: Appointment): boolean {
      return (
        appt.status === APPOINTMENT_STATUS.WITHDRAWN ||
        appt.status === APPOINTMENT_STATUS.CANCELLED
      );
    },
  },
});
