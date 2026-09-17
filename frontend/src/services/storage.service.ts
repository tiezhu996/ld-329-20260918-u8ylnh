import type { Overview } from '../types/domain';

const API_BASE = '/api';

export async function fetchOverview(): Promise<Overview> {
  const response = await fetch(`${API_BASE}/dashboard/overview`);
  if (!response.ok) {
    throw new Error('无法加载校园技能交换数据');
  }
  return response.json() as Promise<Overview>;
}
