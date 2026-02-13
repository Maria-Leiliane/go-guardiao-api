import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { HabitListComponent } from './habit-list/habit-list.component';
import { HabitFormComponent } from './habit-form/habit-form.component';
import { HabitDetailComponent } from './habit-detail/habit-detail.component';

const routes: Routes = [
  { path: '', component: HabitListComponent },
  { path: 'new', component: HabitFormComponent },
  { path: 'edit/:id', component: HabitFormComponent },
  { path: ':id', component: HabitDetailComponent }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class HabitsRoutingModule { }
