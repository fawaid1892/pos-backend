# Sprint Scaffold Instructions

Gunakan template ini pas nge-scaffold sprint baru.

## Cara Pakai

1. Copy `SPRINT-PLAN.md` ke `sprints/sprint-[NUMBER]/`
2. Breakdown task dari PRD / product backlog
3. Assign ke agent yang sesuai
4. Catat di DECISIONS.md kalo ada keputusan arsitektur

## Output per Sprint

Setiap sprint harus menghasilkan:
- `sprints/sprint-[NUMBER]/SPRINT-PLAN.md` — rencana sprint
- `sprints/sprint-[NUMBER]/TASK-LOG.md` — tracking harian
- Update `DECISIONS.md` kalo ada ADR baru
- Update `HEARTBEAT.md` kalo perlu cron job baru

## Delegasi Workflow (via multi-agent-orchestrator skill)

### Create Sprint Plan:
1. PM breakdown task dari PRD → sprint backlog
2. PM panggil `sessions_send` ke tiap agent untuk tiap task
3. PM log semua delegasi di TASK-LOG.md

### Review Task Results:
1. Spawn agent dengan task spesifik (`sessions_spawn`)
2. `sessions_yield` tunggu hasil
3. Review hasil, masukin ke sprint
4. Kalo ada QA task, spawn quality-assurance buat testing