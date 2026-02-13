import { Component, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { HabitService } from '../../../core/services/habit.service';
import { Habit, HabitCompletionHistory } from '../../../core/models/habit.model';

@Component({
  selector: 'app-habit-detail',
  templateUrl: './habit-detail.component.html',
  styleUrls: ['./habit-detail.component.scss']
})
export class HabitDetailComponent implements OnInit {
  habit: Habit | null = null;
  history: HabitCompletionHistory[] = [];
  isLoading = true;

  constructor(
    private habitService: HabitService,
    private route: ActivatedRoute,
    private router: Router
  ) {}

  ngOnInit(): void {
    const habitId = this.route.snapshot.paramMap.get('id');
    if (habitId) {
      this.loadHabit(habitId);
      this.loadHistory(habitId);
    }
  }

  loadHabit(id: string): void {
    this.habitService.getHabitById(id).subscribe({
      next: (habit) => {
        this.habit = habit;
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
      }
    });
  }

  loadHistory(id: string): void {
    this.habitService.getHabitHistory(id).subscribe({
      next: (history) => {
        this.history = history;
      },
      error: () => {
        this.history = [];
      }
    });
  }

  onEdit(): void {
    if (this.habit) {
      this.router.navigate(['/habits/edit', this.habit.id]);
    }
  }

  onBack(): void {
    this.router.navigate(['/habits']);
  }
}
