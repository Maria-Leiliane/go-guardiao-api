export interface LeaderboardEntry {
  rank: number;
  userId: string;
  userName: string;
  mana: number;
  level: number;
  avatar?: string;
}

export interface LeaderboardResponse {
  entries: LeaderboardEntry[];
  userRank?: number;
  totalUsers: number;
}
