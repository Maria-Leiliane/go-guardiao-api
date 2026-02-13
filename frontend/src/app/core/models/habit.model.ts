export interface Habit {
  id: string;
  userId: string;
  name: string;
  description: string;
  frequency: HabitFrequency;
  completed: boolean;
  completionCount: number;
  streak: number;
  createdAt: Date;
  updatedAt: Date;
}

export enum HabitFrequency {
  DAILY = 'daily',
  WEEKLY = 'weekly',
  MONTHLY = 'monthly'
}

export interface HabitCreateRequest {
  name: string;
  description: string;
  frequency: HabitFrequency;
}

export interface HabitUpdateRequest {
  name?: string;
  description?: string;
  frequency?: HabitFrequency;
}

export interface HabitCompletionHistory {
  id: string;
  habitId: string;
  completedAt: Date;
  manaAwarded: number;
}
