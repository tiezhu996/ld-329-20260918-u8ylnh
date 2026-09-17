// 交换预约状态机：与后端 constants/appointment.go 保持一致。
export const APPOINTMENT_STATUS = {
  PENDING: 'pending',
  CONFIRMED_PARTIAL: 'confirmed_partial',
  CONFIRMED: 'confirmed',
  WITHDRAWN: 'withdrawn',
  CANCEL_REQUESTED: 'cancel_requested',
  CANCELLED: 'cancelled',
} as const;

export type AppointmentStatus =
  (typeof APPOINTMENT_STATUS)[keyof typeof APPOINTMENT_STATUS];

export const CANCEL_DECISION = {
  APPROVE: 'approve',
  REJECT: 'reject',
} as const;

export type CancelDecision =
  (typeof CANCEL_DECISION)[keyof typeof CANCEL_DECISION];

export const ROLE = {
  INITIATOR: 'initiator',
  RESPONDER: 'responder',
} as const;

// 终态：重复请求不能改变，刷新后回读仍是该状态。
export const TERMINAL_STATUSES: ReadonlySet<AppointmentStatus> = new Set([
  APPOINTMENT_STATUS.WITHDRAWN,
  APPOINTMENT_STATUS.CANCELLED,
]);

export const STATUS_TAG_TYPE: Record<
  AppointmentStatus,
  'info' | 'warning' | 'success' | 'danger' | 'primary'
> = {
  [APPOINTMENT_STATUS.PENDING]: 'info',
  [APPOINTMENT_STATUS.CONFIRMED_PARTIAL]: 'warning',
  [APPOINTMENT_STATUS.CONFIRMED]: 'success',
  [APPOINTMENT_STATUS.WITHDRAWN]: 'danger',
  [APPOINTMENT_STATUS.CANCEL_REQUESTED]: 'warning',
  [APPOINTMENT_STATUS.CANCELLED]: 'danger',
};

export const MIN_CANCEL_REASON_LENGTH = 4;
export const MAX_CANCEL_REASON_LENGTH = 300;

// 演示用登录用户：JWT 认证预留，当前用 X-User-Name 头传递操作人。
export const DEMO_ACTORS = ['林澈', '孟野', '周芮', '许安'] as const;
