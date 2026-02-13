import { Component, OnInit } from '@angular/core';
import { Router } from '@angular/router';
import { HabitService } from '../../../core/services/habit.service';
import { Habit } from '../../../core/models/habit.model';

@Component({
  selector: 'app-habit-list',
  templateUrl: './habit-list.component.html',
  styleUrls: ['./habit-list.component.scss']
})
export class HabitListComponent implements OnInit {
  habits: Habit[] = [];
  isLoading = true;
  deleteModalOpen = false;
  habitToDelete: string | null = null;

  constructor(
    private habitService: HabitService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.loadHabits();
  }

  loadHabits(): void {
    this.isLoading = true;
    this.habitService.getHabits().subscribe({
      next: (habits) => {
        this.habits = habits;
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
      }
    });
  }

  onComplete(habitId: string): void {
    this.habitService.completeHabit(habitId).subscribe({
      next: () => {
        this.loadHabits();
      }
    });
  }

  onEdit(habitId: string): void {
    this.router.navigate(['/habits/edit', habitId]);
  }

  onDelete(habitId: string): void {
    this.habitToDelete = habitId;
    this.deleteModalOpen = true;
  }

  confirmDelete(): void {
    if (this.habitToDelete) {
      this.habitService.deleteHabit(this.habitToDelete).subscribe({
        next: () => {
          this.deleteModalOpen = false;
          this.habitToDelete = null;
          this.loadHabits();
        }
      });
    }
  }

  cancelDelete(): void {
    this.deleteModalOpen = false;
    this.habitToDelete = null;
  }

  viewDetail(habitId: string): void {
    this.router.navigate(['/habits', habitId]);
  }

  createHabit(): void {
    this.router.navigate(['/habits/new']);
  }
}
