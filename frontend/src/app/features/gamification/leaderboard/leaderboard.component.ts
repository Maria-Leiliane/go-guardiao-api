import { Component, OnInit } from '@angular/core';
import { GamificationService } from '../../../core/services/gamification.service';
import { LeaderboardEntry } from '../../../core/models/leaderboard.model';

@Component({
  selector: 'app-leaderboard',
  templateUrl: './leaderboard.component.html',
  styleUrls: ['./leaderboard.component.scss']
})
export class LeaderboardComponent implements OnInit {
  entries: LeaderboardEntry[] = [];
  userRank: number | undefined;
  isLoading = true;

  constructor(private gamificationService: GamificationService) {}

  ngOnInit(): void {
    this.loadLeaderboard();
  }

  loadLeaderboard(): void {
    this.isLoading = true;
    this.gamificationService.getLeaderboard(100).subscribe({
      next: (response) => {
        this.entries = response.entries;
        this.userRank = response.userRank;
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
      }
    });
  }

  getRankClass(rank: number): string {
    if (rank === 1) return 'gold';
    if (rank === 2) return 'silver';
    if (rank === 3) return 'bronze';
    return '';
  }

  getRankIcon(rank: number): string {
    if (rank === 1) return '🥇';
    if (rank === 2) return '🥈';
    if (rank === 3) return '🥉';
    return '';
  }
}
