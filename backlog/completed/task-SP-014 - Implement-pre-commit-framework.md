---
id: SP-014
title: Implement pre-commit framework
status: Done
assignee: []
created_date: '2025-12-24 15:30'
updated_date: '2025-12-24 15:30'
labels: []
dependencies: []
---

# Description

Migrate from custom shell script hooks to the standard [pre-commit](https://pre-commit.com/) framework.

This allows for:
- Standardized hooks (trailing whitespace, EOF fixer).
- Integration with `golangci-lint`.
- Easier management of local hooks (like our backlog linter).

# Acceptance Criteria

- [x] `.pre-commit-config.yaml` created.
- [x] Standard hooks configured.
- [x] `golangci-lint` hook configured.
- [x] Local `backlog lint` hook configured.
- [x] Documentation updated.
