# 🤝 Contribuindo com o Guardião da Saúde

Olá! Obrigado por se interessar em contribuir com o **Guardião da Saúde**. Buscamos qualidade, robustez e colaboração aberta.  
Leia este guia para entender o fluxo e boas práticas do projeto.

---

## 📝 Como contribuir

1. **Fork este repositório** e crie sua branch:
   ```bash
   git checkout -b minha-feature
   ```

2. **Faça commits claros e pequenos**  
   Use [Conventional Commits](https://www.conventionalcommits.org/) para facilitar o versionamento semântico e o changelog.

3. **Atualize/adicione testes** quando necessário.

4. **Atualize a documentação** (README.md, comentários, exemplos) sempre que alterar funcionalidades públicas.

5. **Garanta que a build e os testes passem**:
   ```bash
   make test
   ```

6. **Abra um Pull Request (PR)**  
   Descreva claramente o que foi feito, motivo da mudança e se houve breaking changes.

---

## 🚦 Fluxo de Pull Requests

- PRs serão revisados por pelo menos um maintainer.
- PRs só são aceitos se a pipeline (build/testes/linters) estiver verde.
- Feedbacks e sugestões podem ser feitos antes do merge.
- PRs com breaking changes devem ser sinalizados.

---

## 🏷️ Convenção de Commits

Adote o padrão **Conventional Commits**. Exemplos:
- `feat: adiciona endpoint de criação de hábito`
- `fix: corrige bug no cálculo de mana`
- `docs: atualiza documentação da API`
- `refactor: melhora performance do worker`
- `test: adiciona testes para autenticação`

---

## 🧑‍💻 Boas práticas de código

- Siga o [Effective Go](https://go.dev/doc/effective_go)
- Use linters (`golangci-lint` recomendado)
- Escreva testes unitários para novas funcionalidades
- Prefira funções pequenas e código legível

---

## 💡 Sugestões e Issues

- Use a aba **Issues** para reportar bugs ou sugerir melhorias.
- Procure uma issue existente antes de abrir uma nova.

---

## 📦 Ambiente de desenvolvimento

- Siga o README para subir o ambiente local com Docker/Makefile.
- Use `.env.development` para variáveis de ambiente locais.
- Evite subir secrets reais para o repositório.

---

## 📜 Licença

Ao contribuir, você concorda que seu código será licenciado sob a licença do projeto (MIT).

---

Obrigado por contribuir com o Guardião da Saúde! 💙