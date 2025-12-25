# Copilot Instructions (Repository Constitution)

This repository follows these rules:

## Backlog-first
- We use Backlog.md: https://github.com/MrLesk/Backlog.md
- For any new work, first create or update a task in `backlog.md`.
- Each task must include acceptance criteria and a test plan.

## Pre-commit is mandatory
- Pre-commit hooks must be enabled and must pass.
- Never suggest disabling hooks, skipping hooks, or using `--no-verify`.
- Fix the underlying issue if hooks fail.

## Tests are mandatory
- All changes require tests (add or update).
- Prefer fast, deterministic unit tests.
- If a change is hard to test, explain why and add an alternative safeguard.

## Output format
For each request, respond in this order:
1) Backlog task update
2) Implementation plan
3) Code changes
4) Tests
5) Pre-commit considerations
6) Documentation updates
- Use markdown with YAML front matter for backlog tasks.
- Use code blocks for code snippets.
- Use bullet points and numbered lists for clarity.
- Do not include explanations or justifications unless explicitly requested.
- Ensure all code snippets are complete and syntactically correct.
## Communication
- Be concise and to the point.
- Use clear and unambiguous language.
- Avoid jargon unless necessary.
- When asking questions, be specific about what information is needed.
## Review process
- Always consider how changes will be reviewed.
- Ensure code is clean, well-documented, and adheres to style guidelines.
- Anticipate potential questions or concerns from reviewers and address them proactively.
## Documentation
- Update relevant documentation for any changes made.
- Ensure documentation is clear, accurate, and easy to understand.
- Use examples where appropriate to illustrate concepts.
## Ethical considerations
- Ensure all suggestions and changes adhere to ethical guidelines.
- Avoid suggestions that could lead to harm, bias, or unethical behavior.
- Prioritize user privacy and data security in all implementations.
## Continuous improvement
- Always look for ways to improve code quality, performance, and maintainability.
- Suggest refactoring or optimization when appropriate.
- Stay updated with best practices and incorporate them into the codebase.
---
