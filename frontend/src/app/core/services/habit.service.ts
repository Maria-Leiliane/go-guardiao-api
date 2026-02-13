import { Injectable } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable } from 'rxjs';
import { environment } from '../../../environments/environment';
import { 
  Habit, 
  HabitCreateRequest, 
  HabitUpdateRequest,
  HabitCompletionHistory 
} from '../models/habit.model';

@Injectable({
  providedIn: 'root'
})
export class HabitService {
  private apiUrl = `${environment.apiUrl}/habits`;

  constructor(private http: HttpClient) {}

  getHabits(): Observable<Habit[]> {
    return this.http.get<Habit[]>(this.apiUrl);
  }

  getHabitById(id: string): Observable<Habit> {
    return this.http.get<Habit>(`${this.apiUrl}/${id}`);
  }

  createHabit(habit: HabitCreateRequest): Observable<Habit> {
    return this.http.post<Habit>(this.apiUrl, habit);
  }

  updateHabit(id: string, habit: HabitUpdateRequest): Observable<Habit> {
    return this.http.put<Habit>(`${this.apiUrl}/${id}`, habit);
  }

  deleteHabit(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`);
  }

  completeHabit(id: string): Observable<any> {
    return this.http.post<any>(`${this.apiUrl}/${id}/complete`, {});
  }

  getHabitHistory(id: string): Observable<HabitCompletionHistory[]> {
    return this.http.get<HabitCompletionHistory[]>(`${this.apiUrl}/${id}/history`);
  }

  getTodayHabits(): Observable<Habit[]> {
    return this.http.get<Habit[]>(`${this.apiUrl}/today`);
  }
}
