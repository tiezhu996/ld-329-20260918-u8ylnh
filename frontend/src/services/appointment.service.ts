import { AppException } from '../errors/AppException';
import type { ApiErrorBody, Appointment, CancelAction } from '../types/appointment';

const API_BASE = '/api';

async function parseError(response: Response): Promise<AppException> {
  try {
    const body = (await response.json()) as ApiErrorBody;
    return new AppException(body.message || '请求失败', body.code || 'REQUEST_FAILED');
  } catch {
    return new AppException('请求失败，请稍后重试', 'REQUEST_FAILED');
  }
}

async function post<T>(url: string, body: unknown): Promise<T> {
  const response = await fetch(url, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  if (!response.ok) {
    throw await parseError(response);
  }
  return response.json() as Promise<T>;
}

export async function fetchAppointments(): Promise<Appointment[]> {
  const response = await fetch(`${API_BASE}/appointments`);
  if (!response.ok) {
    throw await parseError(response);
  }
  return response.json() as Promise<Appointment[]>;
}

export function confirmAppointment(id: number, actor: string): Promise<Appointment> {
  return post<Appointment>(`${API_BASE}/appointments/${id}/confirm`, { actor });
}

export function withdrawAppointment(id: number, actor: string): Promise<Appointment> {
  return post<Appointment>(`${API_BASE}/appointments/${id}/withdraw`, { actor });
}

export function requestCancel(id: number, actor: string, reason: string): Promise<Appointment> {
  return post<Appointment>(`${API_BASE}/appointments/${id}/cancel-requests`, { actor, reason });
}

export function respondCancel(id: number, actor: string, action: CancelAction): Promise<Appointment> {
  return post<Appointment>(`${API_BASE}/appointments/${id}/cancel-requests/respond`, { actor, action });
}
