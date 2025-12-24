---
id: SP-013
title: Optimize sync and enrich commands for idempotency
status: Done
assignee: []
created_date: '2025-12-24 15:00'
updated_date: '2025-12-24 15:00'
labels: []
dependencies: []
---

# Description

Optimize `sync` and `enrich` commands to be efficient and idempotent.

- `sync`: Compare fetched games with existing DB records. Only update if changed. Report stats (New/Updated/Unchanged).
- `enrich`: Skip games that already have details in the DB. Add `--refresh` flag to force update.

# Acceptance Criteria

- [x] `sync` command reports statistics.
- [x] `enrich` command skips existing details by default.
- [x] `enrich` command supports `--refresh` to force update.
