export interface Challenge {
  id: string;
  title: string;
  description: string;
  reward: number;
  status: ChallengeStatus;
  progress: number;
  targetProgress: number;
  expiresAt?: Date;
  createdAt: Date;
}

export enum ChallengeStatus {
  ACTIVE = 'active',
  COMPLETED = 'completed',
  EXPIRED = 'expired'
}

export interface ChallengeProgress {
  challengeId: string;
  progress: number;
  completed: boolean;
}
