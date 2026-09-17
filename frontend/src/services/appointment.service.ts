import { AppException } from '../errors/AppException';
import { logger } from '../logger/logger';
import type {
  Appointment,
  AppointmentActionResult,
  CreateAppointmentPayload,
} from '../types/domain';
import type { CancelDecision } from '../constants/appointment.constants';

const API_BASE = '/api';
const ACTOR_HEADER = 'X-User-Name';

interface BackendErrorBody {
  error?: { code: string; message: string };
}

async function sendAction(
  path: string,
  actor: string,
  body?: unknown,
): Promise<AppointmentActionResult> {
  const response = await fetch(`${API_BASE}${path}`, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      [ACTOR_HEADER]: actor,
    },
    body: body === undefined ? undefined : JSON.stringify(body),
  });
  return parseActionResponse(response);
}

async function parseActionResponse(
  response: Response,
): Promise<AppointmentActionResult> {
  const data = (await response.json().catch(() => null)) as
    | (AppointmentActionResult & BackendErrorBody)
    | null;
  if (!response.ok || !data) {
    const message = data?.error?.message ?? '预约操作失败，请稍后重试';
    logger.warn('appointment action rejected', response.status, data);
    throw new AppException(message, data?.error?.code ?? 'ACTION_FAILED');
  }
  return data as AppointmentActionResult;
}

export const appointmentApi = {
  async list(): Promise<Appointment[]> {
    const response = await fetch(`${API_BASE}/appointments`);
    if (!response.ok) {
      throw new AppException('无法加载交换预约');
    }
    const data = (await response.json()) as { appointments: Appointment[] };
    return data.appointments;
  },

  create(payload: CreateAppointmentPayload): Promise<AppointmentActionResult> {
    // 创建接口返回的是 { appointment }，包一层与操作结果保持同构。
    return fetch(`${API_BASE}/appointments`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json', [ACTOR_HEADER]: payload.initiator },
      body: JSON.stringify(payload),
    }).then(async (response) => {
      const data = (await response.json().catch(() => null)) as
        | ({ appointment: Appointment } & BackendErrorBody)
        | null;
      if (!response.ok || !data) {
        throw new AppException(
          data?.error?.message ?? '发起预约失败',
          data?.error?.code ?? 'CREATE_FAILED',
        );
      }
      return { appointment: data.appointment, changed: true };
    });
  },

  confirm(id: number, actor: string) {
    return sendAction(`/appointments/${id}/confirm`, actor);
  },

  withdraw(id: number, actor: string) {
    return sendAction(`/appointments/${id}/withdraw`, actor);
  },

  requestCancel(id: number, actor: string, reason: string) {
    return sendAction(`/appointments/${id}/cancel-requests`, actor, { reason });
  },

  decideCancel(id: number, actor: string, decision: CancelDecision) {
    return sendAction(`/appointments/${id}/cancel-decisions`, actor, { decision });
  },
};
