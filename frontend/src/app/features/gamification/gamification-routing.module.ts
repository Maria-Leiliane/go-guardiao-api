import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';
import { ManaComponent } from './mana/mana.component';
import { LeaderboardComponent } from './leaderboard/leaderboard.component';
import { ChallengesComponent } from './challenges/challenges.component';

const routes: Routes = [
  { path: '', redirectTo: 'mana', pathMatch: 'full' },
  { path: 'mana', component: ManaComponent },
  { path: 'leaderboard', component: LeaderboardComponent },
  { path: 'challenges', component: ChallengesComponent }
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class GamificationRoutingModule { }
