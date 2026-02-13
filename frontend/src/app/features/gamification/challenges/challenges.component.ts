import { Component, OnInit } from '@angular/core';
import { GamificationService } from '../../../core/services/gamification.service';
import { Challenge, ChallengeStatus } from '../../../core/models/challenge.model';

@Component({
  selector: 'app-challenges',
  templateUrl: './challenges.component.html',
  styleUrls: ['./challenges.component.scss']
})
export class ChallengesComponent implements OnInit {
  challenges: Challenge[] = [];
  isLoading = true;
  filter: 'all' | 'active' | 'completed' = 'all';

  constructor(private gamificationService: GamificationService) {}

  ngOnInit(): void {
    this.loadChallenges();
  }

  loadChallenges(): void {
    this.isLoading = true;
    this.gamificationService.getChallenges().subscribe({
      next: (challenges) => {
        this.challenges = challenges;
        this.isLoading = false;
      },
      error: () => {
        this.isLoading = false;
      }
    });
  }

  get filteredChallenges(): Challenge[] {
    if (this.filter === 'all') return this.challenges;
    if (this.filter === 'active') {
      return this.challenges.filter(c => c.status === ChallengeStatus.ACTIVE);
    }
    return this.challenges.filter(c => c.status === ChallengeStatus.COMPLETED);
  }

  getProgressPercentage(challenge: Challenge): number {
    return (challenge.progress / challenge.targetProgress) * 100;
  }

  getStatusLabel(status: ChallengeStatus): string {
    const labels = {
      [ChallengeStatus.ACTIVE]: 'Ativo',
      [ChallengeStatus.COMPLETED]: 'Completo',
      [ChallengeStatus.EXPIRED]: 'Expirado'
    };
    return labels[status];
  }

  getStatusClass(status: ChallengeStatus): string {
    const classes = {
      [ChallengeStatus.ACTIVE]: 'status-active',
      [ChallengeStatus.COMPLETED]: 'status-completed',
      [ChallengeStatus.EXPIRED]: 'status-expired'
    };
    return classes[status];
  }
}
