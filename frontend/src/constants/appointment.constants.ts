import type { AppointmentState } from '../types/appointment';

export const APPOINTMENT_STATE_TAG: Record<AppointmentState, 'primary' | 'success' | 'warning' | 'info' | 'danger'> = {
  pending: 'warning',
  confirmed: 'success',
  cancel_pending: 'danger',
  withdrawn: 'info',
  cancelled: 'info',
};

export const APPOINTMENT_PANEL_TITLE = '预约确认';
export const SLOT_RELEASED_TEXT = '双方时间档已释放';
export const WITHDRAW_CONFIRM_TEXT = '撤回后预约立即失效，双方时间档将释放，确定撤回吗？';
export const CANCEL_DIALOG_TITLE = '申请取消预约';
export const CANCEL_REASON_PLACEHOLDER = '请填写取消原因，提交后需等待对方处理';
export const CANCEL_REASON_REQUIRED_TEXT = '请填写取消原因';
