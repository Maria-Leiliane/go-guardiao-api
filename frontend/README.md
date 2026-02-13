# Go Guardião - Frontend

## Descrição

Single Page Application (SPA) do Go Guardião, uma plataforma de gerenciamento de hábitos com gamificação para apoio oncológico.

## Tecnologias

- **Framework:** Angular 17
- **Linguagem:** TypeScript
- **Estilização:** SCSS
- **HTTP Client:** RxJS e HttpClient
- **Autenticação:** JWT com HTTP Interceptors

## Estrutura do Projeto

```
frontend/
├── src/
│   ├── app/
│   │   ├── core/                 # Serviços, guards, interceptors, models
│   │   ├── features/             # Módulos de features (auth, dashboard, habits, etc)
│   │   └── shared/               # Componentes compartilhados
│   ├── environments/             # Configurações de ambiente
│   └── styles.scss               # Estilos globais
```

## Instalação

1. Certifique-se de ter o Node.js instalado (versão 18 ou superior)
2. Entre na pasta frontend:
   ```bash
   cd frontend
   ```
3. Instale as dependências:
   ```bash
   npm install
   ```

## Comandos Disponíveis

- **Iniciar servidor de desenvolvimento:**
  ```bash
  npm start
  ```
  Acesse `http://localhost:4200`

- **Build para produção:**
  ```bash
  npm run build
  ```
  Os arquivos compilados estarão em `dist/`

- **Executar testes:**
  ```bash
  npm test
  ```

- **Executar linter:**
  ```bash
  npm run lint
  ```

## Funcionalidades

### 1. Autenticação
- Login com email e senha
- Registro de novos usuários
- Logout
- Proteção de rotas com AuthGuard
- JWT token gerenciado automaticamente via interceptor

### 2. Dashboard
- Visão geral do progresso do usuário
- Mana atual e barra de progresso
- Hábitos do dia
- Desafios ativos
- Estatísticas rápidas

### 3. Hábitos
- Criar, editar e excluir hábitos
- Marcar hábitos como completos
- Visualizar histórico de conclusões
- Frequências: diário, semanal, mensal
- Sistema de sequência (streak)

### 4. Perfil
- Visualizar e editar dados pessoais
- Ver nível e Mana total
- Contatos de suporte oncológico

### 5. Gamificação
- **Mana:** Sistema de pontos com níveis
- **Leaderboard:** Ranking de usuários
- **Desafios:** Lista de desafios com recompensas

## Configuração da API

A URL da API pode ser configurada em:
- `src/environments/environment.ts` (desenvolvimento)
- `src/environments/environment.prod.ts` (produção)

Por padrão, a API está configurada para `http://localhost:8080/api`

## Arquitetura

### Core Module
Contém a lógica de negócio principal:
- **Services:** Comunicação com a API
- **Guards:** Proteção de rotas
- **Interceptors:** Manipulação de requisições HTTP
- **Models:** Interfaces TypeScript

### Feature Modules
Módulos lazy-loaded para cada funcionalidade principal:
- AuthModule
- DashboardModule
- HabitsModule
- ProfileModule
- GamificationModule

### Shared Module
Componentes reutilizáveis em toda a aplicação:
- Navbar
- Card
- Modal
- Button

## Estilização

O projeto usa variáveis CSS para temas consistentes:
- Cores da marca (verde/azul)
- Sistema de espaçamento
- Tipografia responsiva
- Design mobile-first

## Autenticação

O fluxo de autenticação:
1. Login/Registro via AuthService
2. JWT token armazenado no localStorage
3. AuthInterceptor adiciona token em todas as requisições
4. AuthGuard protege rotas autenticadas
5. Logout limpa token e redireciona para login

## Contribuindo

1. Crie uma branch para sua feature
2. Faça suas alterações
3. Teste localmente
4. Envie um pull request

## Licença

Este projeto está sob a licença MIT.
