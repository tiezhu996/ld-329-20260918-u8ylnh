import { defineStore } from 'pinia';
import { sampleSkills } from '../data/sample.data';

export const useDashboardStore = defineStore('dashboard', {
  state: () => ({ skills: sampleSkills, keyword: '' }),
});
