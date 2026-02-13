import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { HabitService } from '../../../core/services/habit.service';
import { HabitFrequency } from '../../../core/models/habit.model';

@Component({
  selector: 'app-habit-form',
  templateUrl: './habit-form.component.html',
  styleUrls: ['./habit-form.component.scss']
})
export class HabitFormComponent implements OnInit {
  habitForm: FormGroup;
  isEditMode = false;
  habitId: string | null = null;
  isLoading = false;
  errorMessage = '';
  frequencies = [
    { value: HabitFrequency.DAILY, label: 'Diário' },
    { value: HabitFrequency.WEEKLY, label: 'Semanal' },
    { value: HabitFrequency.MONTHLY, label: 'Mensal' }
  ];

  constructor(
    private fb: FormBuilder,
    private habitService: HabitService,
    private router: Router,
    private route: ActivatedRoute
  ) {
    this.habitForm = this.fb.group({
      name: ['', [Validators.required, Validators.minLength(3)]],
      description: ['', [Validators.required]],
      frequency: [HabitFrequency.DAILY, [Validators.required]]
    });
  }

  ngOnInit(): void {
    this.habitId = this.route.snapshot.paramMap.get('id');
    if (this.habitId) {
      this.isEditMode = true;
      this.loadHabit();
    }
  }

  loadHabit(): void {
    if (this.habitId) {
      this.habitService.getHabitById(this.habitId).subscribe({
        next: (habit) => {
          this.habitForm.patchValue({
            name: habit.name,
            description: habit.description,
            frequency: habit.frequency
          });
        }
      });
    }
  }

  onSubmit(): void {
    if (this.habitForm.valid) {
      this.isLoading = true;
      this.errorMessage = '';

      const habitData = this.habitForm.value;

      if (this.isEditMode && this.habitId) {
        this.habitService.updateHabit(this.habitId, habitData).subscribe({
          next: () => {
            this.router.navigate(['/habits']);
          },
          error: (error) => {
            this.isLoading = false;
            this.errorMessage = error.error?.message || 'Erro ao atualizar hábito.';
          }
        });
      } else {
        this.habitService.createHabit(habitData).subscribe({
          next: () => {
            this.router.navigate(['/habits']);
          },
          error: (error) => {
            this.isLoading = false;
            this.errorMessage = error.error?.message || 'Erro ao criar hábito.';
          }
        });
      }
    }
  }

  onCancel(): void {
    this.router.navigate(['/habits']);
  }
}
