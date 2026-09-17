import type { AppointmentStatus, CancelDecision } from '../constants/appointment.constants';

export interface Skill {
  id: number;
  owner: string;
  title: string;
  category: string;
  level: number;
  campus: string;
  description: string;
  timeSlots: string[];
  rewards: string[];
  portfolio: string;
}

export interface Need {
  id: number;
  requester: string;
  title: string;
  category: string;
  campus: string;
  expectTime: string;
  budgetType: string;
  description: string;
  responses: number;
}

export interface Match {
  id: number;
  provider: string;
  learner: string;
  offerSkill: string;
  wantedSkill: string;
  score: number;
  commonSlots: string[];
  recommendation: string;
}

export interface CancelRequest {
  by: string;
  reason: string;
  at: number;
  decidedAt?: number;
  decision?: CancelDecision;
}

export interface Appointment {
  id: number;
  initiator: string;
  responder: string;
  pair: string;
  time: string;
  place: string;
  agenda: string;
  status: AppointmentStatus;
  statusText: string;
  initiatorConfirmed: boolean;
  responderConfirmed: boolean;
  slots: string[];
  slotsHeld: boolean;
  cancel?: CancelRequest;
  version: number;
  createdAt: number;
  updatedAt: number;
}

export interface AppointmentActionResult {
  appointment: Appointment;
  changed: boolean;
  slotsFreed?: string[];
  note?: string;
}

export interface CreateAppointmentPayload {
  initiator: string;
  responder: string;
  time: string;
  place: string;
  agenda: string;
  slots: string[];
}

export interface Review {
  id: number;
  from: string;
  to: string;
  rating: number;
  content: string;
}

export interface Conversation {
  id: number;
  withUser: string;
  unread: number;
  messages: string[];
}

export interface Profile {
  name: string;
  major: string;
  creditScore: number;
  creditLevel: string;
  skillWall: Skill[];
  radar: Record<string, number>;
  history: string[];
  reviews: Review[];
}

export interface Overview {
  service: string;
  categories: string[];
  metrics: Record<string, number>;
  skills: Skill[];
  needs: Need[];
  matches: Match[];
  appointments: Appointment[];
  reviews: Review[];
  messages: Conversation[];
  profile: Profile;
}
