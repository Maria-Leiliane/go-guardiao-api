import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ReactiveFormsModule } from '@angular/forms';
import { HabitsRoutingModule } from './habits-routing.module';
import { SharedModule } from '../../shared/shared.module';
import { HabitListComponent } from './habit-list/habit-list.component';
import { HabitFormComponent } from './habit-form/habit-form.component';
import { HabitDetailComponent } from './habit-detail/habit-detail.component';

@NgModule({
  declarations: [
    HabitListComponent,
    HabitFormComponent,
    HabitDetailComponent
  ],
  imports: [
    CommonModule,
    ReactiveFormsModule,
    HabitsRoutingModule,
    SharedModule
  ]
})
export class HabitsModule { }
