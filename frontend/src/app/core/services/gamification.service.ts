import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Challenge, ChallengeProgress } from '../models/challenge.model';
import { LeaderboardResponse } from '../models/leaderboard.model';

@Injectable({
  providedIn: 'root'
})
export class GamificationService {
  private apiUrl = environment.apiUrl;

  constructor(private http: HttpClient) {}

  // Mana operations
  getManaInfo(): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/gamification/mana`);
  }

  // Challenge operations
  getChallenges(): Observable<Challenge[]> {
    return this.http.get<Challenge[]>(`${this.apiUrl}/gamification/challenges`);
  }

  getActiveChallenges(): Observable<Challenge[]> {
    return this.http.get<Challenge[]>(`${this.apiUrl}/gamification/challenges/active`);
  }

  getChallengeById(id: string): Observable<Challenge> {
    return this.http.get<Challenge>(`${this.apiUrl}/gamification/challenges/${id}`);
  }

  updateChallengeProgress(id: string, progress: number): Observable<ChallengeProgress> {
    return this.http.post<ChallengeProgress>(
      `${this.apiUrl}/gamification/challenges/${id}/progress`, 
      { progress }
    );
  }

  // Leaderboard operations
  getLeaderboard(limit: number = 100): Observable<LeaderboardResponse> {
    return this.http.get<LeaderboardResponse>(`${this.apiUrl}/gamification/leaderboard?limit=${limit}`);
  }

  getUserRank(): Observable<any> {
    return this.http.get<any>(`${this.apiUrl}/gamification/rank`);
  }
}
