import { Component, OnInit } from '@angular/core';
import { GamificationService } from '../../../core/services/gamification.service';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-mana',
  templateUrl: './mana.component.html',
  styleUrls: ['./mana.component.scss']
})
export class ManaComponent implements OnInit {
  manaInfo: any = { current: 0, nextLevel: 100, level: 1, percentage: 0 };
  isLoading = true;

  constructor(
    private gamificationService: GamificationService,
    private authService: AuthService
  ) {}

  ngOnInit(): void {
    this.loadManaInfo();
  }

  loadManaInfo(): void {
    this.isLoading = true;
    this.gamificationService.getManaInfo().subscribe({
      next: (info) => {
        this.manaInfo = {
          ...info,
          percentage: (info.current / info.nextLevel) * 100
        };
        this.isLoading = false;
      },
      error: () => {
        const user = this.authService.getCurrentUser();
        this.manaInfo = {
          current: user?.mana || 0,
          nextLevel: 100,
          level: user?.level || 1,
          percentage: ((user?.mana || 0) / 100) * 100
        };
        this.isLoading = false;
      }
    });
  }
}
