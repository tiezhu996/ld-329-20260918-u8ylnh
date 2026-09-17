// 交换预约状态机状态
export type AppointmentState = 'pending' | 'confirmed' | 'cancel_pending' | 'withdrawn' | 'cancelled';

export type CancelRequestStatus = 'pending' | 'accepted' | 'rejected';

export interface CancelRequest {
  by: string;
  reason: string;
  status: CancelRequestStatus;
  createdAt: string;
  handledBy?: string;
  handledAt?: string;
}

export interface Appointment {
  id: number;
  pair: string;
  initiator: string;
  participant: string;
  time: string;
  place: string;
  state: AppointmentState;
  status: string;
  agenda: string;
  confirmations: string[];
  cancelRequest?: CancelRequest;
  slotReleased: boolean;
  releasedAt?: string;
  version: number;
  updatedAt: string;
}

export type CancelAction = 'accept' | 'reject';

export interface ApiErrorBody {
  code: string;
  message: string;
}
