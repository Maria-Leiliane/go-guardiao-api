import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators } from '@angular/forms';
import { AuthService } from '../../../core/services/auth.service';
import { UserService } from '../../../core/services/user.service';
import { User } from '../../../core/models/user.model';

@Component({
  selector: 'app-profile',
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.scss']
})
export class ProfileComponent implements OnInit {
  currentUser: User | null = null;
  profileForm: FormGroup;
  isEditing = false;
  isLoading = false;
  successMessage = '';
  errorMessage = '';

  supportContacts = [
    { name: 'INCA - Instituto Nacional de Câncer', phone: '0800 61 4000', url: 'https://www.inca.gov.br' },
    { name: 'CVV - Centro de Valorização da Vida', phone: '188', url: 'https://www.cvv.org.br' },
    { name: 'Fundação do Câncer', phone: '(21) 3547-3232', url: 'https://www.cancer.org.br' }
  ];

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private userService: UserService
  ) {
    this.profileForm = this.fb.group({
      name: ['', [Validators.required, Validators.minLength(3)]],
      email: ['', [Validators.required, Validators.email]]
    });
  }

  ngOnInit(): void {
    this.loadUserProfile();
  }

  loadUserProfile(): void {
    this.authService.currentUser$.subscribe(user => {
      if (user) {
        this.currentUser = user;
        this.profileForm.patchValue({
          name: user.name,
          email: user.email
        });
      }
    });
  }

  onEdit(): void {
    this.isEditing = true;
    this.successMessage = '';
    this.errorMessage = '';
  }

  onCancel(): void {
    this.isEditing = false;
    this.loadUserProfile();
    this.successMessage = '';
    this.errorMessage = '';
  }

  onSave(): void {
    if (this.profileForm.valid) {
      this.isLoading = true;
      this.errorMessage = '';
      this.successMessage = '';

      this.userService.updateUserProfile(this.profileForm.value).subscribe({
        next: (updatedUser) => {
          this.currentUser = updatedUser;
          localStorage.setItem('user', JSON.stringify(updatedUser));
          this.isEditing = false;
          this.isLoading = false;
          this.successMessage = 'Perfil atualizado com sucesso!';
        },
        error: (error) => {
          this.isLoading = false;
          this.errorMessage = error.error?.message || 'Erro ao atualizar perfil.';
        }
      });
    }
  }
}
