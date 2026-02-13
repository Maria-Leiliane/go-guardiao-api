import { Component, OnInit } from '@angular/core';
import { AuthService } from '../../../core/services/auth.service';
import { HabitService } from '../../../core/services/habit.service';
import { GamificationService } from '../../../core/services/gamification.service';
import { User } from '../../../core/models/user.model';
import { Habit } from '../../../core/models/habit.model';
import { Challenge } from '../../../core/models/challenge.model';

@Component({
  selector: 'app-dashboard',
  templateUrl: './dashboard.component.html',
  styleUrls: ['./dashboard.component.scss']
})
export class DashboardComponent implements OnInit {
  currentUser: User | null = null;
  todayHabits: Habit[] = [];
  activeChallenges: Challenge[] = [];
  manaInfo: any = { current: 0, nextLevel: 100, level: 1 };
  isLoading = true;

  constructor(
    private authService: AuthService,
    private habitService: HabitService,
    private gamificationService: GamificationService
  ) {}

  ngOnInit(): void {
    this.loadDashboardData();
  }

  loadDashboardData(): void {
    this.isLoading = true;

    this.authService.currentUser$.subscribe(user => {
      this.currentUser = user;
    });

    this.habitService.getTodayHabits().subscribe({
      next: (habits) => {
        this.todayHabits = habits;
      },
      error: () => {
        this.todayHabits = [];
      }
    });

    this.gamificationService.getActiveChallenges().subscribe({
      next: (challenges) => {
        this.activeChallenges = challenges;
      },
      error: () => {
        this.activeChallenges = [];
      }
    });

    this.gamificationService.getManaInfo().subscribe({
      next: (info) => {
        this.manaInfo = info;
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
      }
    });
  }

  onHabitComplete(habitId: string): void {
    this.habitService.completeHabit(habitId).subscribe({
      next: () => {
        this.loadDashboardData();
      },
      error: (error) => {
        console.error('Error completing habit:', error);
      }
    });
  }

  getManaPercentage(): number {
    if (this.manaInfo.nextLevel === 0) return 100;
    return (this.manaInfo.current / this.manaInfo.nextLevel) * 100;
  }
}
