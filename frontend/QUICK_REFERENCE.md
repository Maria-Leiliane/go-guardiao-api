# Go Guardião Frontend - Quick Reference Card

## 🚀 Essential Commands

```bash
# Install dependencies
npm install

# Start development server (http://localhost:4200)
npm start

# Build for production
npm run build

# Run tests
npm test

# Run linter
npm run lint
```

## 📂 Key File Locations

```
frontend/src/
├── environments/environment.ts     # API URL configuration
├── styles.scss                     # Global theme & CSS variables
├── app/
│   ├── app-routing.module.ts       # Main routing configuration
│   ├── core/
│   │   ├── services/               # API service implementations
│   │   ├── models/                 # TypeScript interfaces
│   │   ├── guards/auth.guard.ts    # Route protection
│   │   └── interceptors/auth.interceptor.ts  # JWT injection
│   └── features/                   # All feature modules
```

## 🔧 Configuration

### API Endpoint
Edit `src/environments/environment.ts`:
```typescript
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080/api'  // Change this for your API
};
```

### Theme Colors
Edit `src/styles.scss`:
```scss
:root {
  --primary-color: #4CAF50;    // Main green
  --secondary-color: #2196F3;  // Blue
  --accent-color: #FFC107;     // Amber/Yellow
}
```

## 🧭 Routing Structure

```
/auth/login              → Login page
/auth/register           → Registration page
/dashboard               → Main dashboard (protected)
/habits                  → Habit list (protected)
/habits/new              → Create habit (protected)
/habits/edit/:id         → Edit habit (protected)
/habits/:id              → Habit detail (protected)
/profile                 → User profile (protected)
/gamification/mana       → Mana visualization (protected)
/gamification/leaderboard → Leaderboard (protected)
/gamification/challenges  → Challenges list (protected)
```

## 📡 API Service Methods

### AuthService
```typescript
login(credentials)           // POST /api/auth/login
register(data)               // POST /api/auth/register
logout()                     // Clear local storage
isAuthenticated()            // Check auth status
getCurrentUser()             // Get current user from state
```

### HabitService
```typescript
getHabits()                  // GET /api/habits
getHabitById(id)             // GET /api/habits/:id
createHabit(habit)           // POST /api/habits
updateHabit(id, habit)       // PUT /api/habits/:id
deleteHabit(id)              // DELETE /api/habits/:id
completeHabit(id)            // POST /api/habits/:id/complete
getTodayHabits()             // GET /api/habits/today
```

### GamificationService
```typescript
getManaInfo()                // GET /api/gamification/mana
getChallenges()              // GET /api/gamification/challenges
getLeaderboard(limit)        // GET /api/gamification/leaderboard
```

### UserService
```typescript
getUserProfile()             // GET /api/users/profile
updateUserProfile(data)      // PUT /api/users/profile
```

## 🎨 Shared Components Usage

### Button
```html
<app-button 
  variant="primary"          <!-- primary | secondary | danger | success -->
  size="medium"              <!-- small | medium | large -->
  (clicked)="handleClick()">
  Click Me
</app-button>
```

### Card
```html
<app-card title="Card Title" subtitle="Optional subtitle">
  <!-- Card content here -->
  <div card-footer>
    <!-- Footer content (optional) -->
  </div>
</app-card>
```

### Modal
```html
<app-modal 
  [isOpen]="showModal" 
  title="Modal Title"
  (close)="showModal = false">
  <!-- Modal content -->
  <div modal-footer>
    <app-button (clicked)="confirm()">Confirm</app-button>
  </div>
</app-modal>
```

## 🔐 Authentication Flow

1. **Login/Register** → User submits credentials
2. **AuthService** → Makes API call
3. **Response** → JWT token + user data received
4. **Storage** → Token saved to localStorage
5. **Navigation** → Redirect to dashboard
6. **AuthInterceptor** → Automatically adds token to all requests
7. **AuthGuard** → Protects routes, redirects if not authenticated

## 📱 Responsive Breakpoints

```scss
// Mobile
@media (max-width: 768px) {
  // Mobile styles
}

// Desktop
@media (min-width: 769px) {
  // Desktop styles
}
```

## 🎯 Common Tasks

### Add New Feature Module
1. Create module directory in `features/`
2. Generate components with Angular CLI or manually
3. Create routing module with lazy loading
4. Add route to `app-routing.module.ts`

### Add New Service
1. Create service in `core/services/`
2. Use `@Injectable({ providedIn: 'root' })`
3. Inject HttpClient for API calls
4. Use environment.apiUrl for base URL

### Add New Model
1. Create interface in `core/models/`
2. Export interface
3. Import where needed

### Add New Shared Component
1. Create component in `shared/components/`
2. Add to `shared.module.ts` declarations and exports
3. Use in any module that imports SharedModule

## 🐛 Troubleshooting

### Issue: Module not found
```bash
npm install
```

### Issue: Port already in use
```bash
# Use different port
ng serve --port 4201
```

### Issue: API CORS errors
- Ensure backend allows CORS from `http://localhost:4200`
- Check API URL in environment files

### Issue: Authentication not working
- Check localStorage for 'token' key
- Verify AuthInterceptor is registered in app.module.ts
- Check API returns correct JWT format

## 📚 Useful Resources

- [Angular Documentation](https://angular.io/docs)
- [RxJS Documentation](https://rxjs.dev/)
- [TypeScript Handbook](https://www.typescriptlang.org/docs/)
- [SCSS Documentation](https://sass-lang.com/documentation)

## 🎓 Development Tips

1. **Use Angular DevTools** - Chrome extension for debugging
2. **Enable source maps** - Already configured for development
3. **Use reactive forms** - Already implemented throughout
4. **Follow naming conventions** - kebab-case for files, PascalCase for classes
5. **Keep components small** - Single responsibility principle
6. **Use async pipe** - Automatic subscription management

## 📝 Code Style

- **Indentation**: 2 spaces
- **Quotes**: Single quotes for TypeScript, double for HTML
- **Semicolons**: Required
- **Line length**: Keep under 120 characters
- **Naming**: 
  - Components: `feature-name.component.ts`
  - Services: `feature.service.ts`
  - Models: `feature.model.ts`

---

**Need help?** Check the main README.md or IMPLEMENTATION_SUMMARY.md for detailed information.
