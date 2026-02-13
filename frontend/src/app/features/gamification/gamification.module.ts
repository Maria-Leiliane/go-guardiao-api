import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { GamificationRoutingModule } from './gamification-routing.module';
import { SharedModule } from '../../shared/shared.module';
import { ManaComponent } from './mana/mana.component';
import { LeaderboardComponent } from './leaderboard/leaderboard.component';
import { ChallengesComponent } from './challenges/challenges.component';

@NgModule({
  declarations: [
    ManaComponent,
    LeaderboardComponent,
    ChallengesComponent
  ],
  imports: [
    CommonModule,
    GamificationRoutingModule,
    SharedModule
  ]
})
export class GamificationModule { }
