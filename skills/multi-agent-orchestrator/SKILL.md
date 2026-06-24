---
name: "multi-agent-orchestrator"
description: "PM orchestration + sprint management: delegate tasks, daily standup, sprint scaffolding, decision log."
---

# Multi-Agent Orchestrator Skill

## Tujuan
Memberikan kemampuan sebagai **Project Manager (Orchestrator)** untuk mendelegasikan task ke agent lain (frontend-dev, backend-dev, quality-assurance) menggunakan internal OpenClaw tools, serta mengelola sprint lifecycle.

## Agent yang Diorganisir

| Role | Agent ID | Tujuan |
|------|----------|--------|
| 🧑‍💼 PM | `project-manager` | Orchestrator — breakdown task, delegate, track progress |
| 📱 Frontend | `frontend-dev` | UI/UX implementation (Flutter, React, CSS) |
| 🛠️ Backend | `backend-dev` | API, database, server logic (Go, Node.js, Supabase) |
| 🧪 QA | `quality-assurance` | Testing, test cases, bug reports |

## Tool yang Dipakai

### 1. `sessions_send` — Delegasi Langsung
Gunakan saat task sudah jelas dan tinggal dikerjakan.

```
sessions_send(
  agentId: "frontend-dev",
  message: "Instruksi jelas dengan konteks dan output yang diharapkan"
)
```

### 2. `sessions_spawn` — Sub-Agent untuk Task Spesifik
Gunakan saat butuh agent melakukan task terisolasi dan return hasil.

```
sessions_spawn(
  agentId: "backend-dev",
  task: "Deskripsi task lengkap + path dokumen referensi",
  taskName: "nama-unik-task"
)
```

### 3. `sessions_yield` — Tunggu Hasil Sub-Agent
Setelah spawn, panggil yield untuk menunggu hasil dari sub-agent.

```
sessions_yield(message: "Menunggu hasil delegasi...")
```

## Workflow Delegasi

### Flow 1: Delegasi Langsung (Fire & Forget)
1. PM breakdown task dari user
2. PM panggil `sessions_send` ke agent yang tepat dengan instruksi lengkap
3. PM informasikan ke user bahwa task sudah didelegasikan

### Flow 2: Delegasi dengan Tunggu Hasil (Fire & Wait)
1. PM breakdown task
2. PM spawn sub-agent dengan task spesifik
3. PM panggil `sessions_yield` untuk menunggu
4. Setelah hasil masuk, PM compile dan laporkan ke user
5. Jika task kompleks, PM bisa spawn beberapa sub-agent bergantian

### Flow 3: Diskusi Tim (War Room)
1. User kasih request fitur baru
2. PM breakdown requirements
3. Step 1: PM spawn frontend-dev untuk desain UI + state management
4. Step 2: PM spawn backend-dev untuk desain API + DB schema
5. Step 3: PM spawn quality-assurance untuk test case
6. Setelah semua hasil terkumpul, PM compile jadi 1 dokumen laporan

### Flow 4: Sprint Scaffold
1. User minta buka sprint baru
2. PM breakdown task dari PRD / product backlog jadi sprint backlog
3. PM buat folder `sprints/sprint-[NUMBER]/`
4. PM generate SPRINT-PLAN.md, TASK-LOG.md, SPRINT-REVIEW.md
5. PM delegasi tiap task ke agent yang sesuai via `sessions_send`
6. PM catat ADR (Architecture Decision Record) baru di `DECISIONS.md` kalo ada

### Flow 5: Daily Standup (via Cron)
Dijadwalkan otomatis via cron job `daily-standup` tiap jam 09:00 WIB (Senin-Sabtu).
1. Cron trigger → isolated session PM
2. PM baca sprint plans & task logs
3. PM cek progress tiap task
4. PM identifikasi blocker
5. PM kirim ringkasan standup ke user di Telegram

### Flow 6: Decision Log (ADR)
Setiap ada keputusan arsitektur penting, catat di `DECISIONS.md`:
1. Tentukan nomor ADR berikutnya
2. Tulis: Date, Status, Context, Decision, Alternatives, Consequences
3. Update status ADR sebelumnya kalo ada yang deprecated/superseded

## Aturan Delegasi

### 1. Instruksi Harus Lengkap
Setiap delegasi WAJIB mengandung:
- **Konteks**: apa yang sedang dikerjakan, link ke dokumen terkait
- **Task spesifik**: apa yang harus dikerjakan
- **Priority**: sprint berapa, urgency
- **Output yang diharapkan**: file apa, format apa
- **Deadline/estimasi**: kalo ada

### 2. API Contract Wajib Disepakati Sebelum Coding
- Kalo task butuh frontend + backend, tunggu backend selesai desain API dulu
- Atau: backend kirim API contract → frontend bisa mock data sambil nunggu
- QA review contract dari sisi edge cases

### 3. Reporting ke User
- Setelah delegasi, informasikan ke user: "Task X sudah didelegasikan ke agent Y"
- Kalo ada hasil, ringkas dalam bahasa Indonesia yang jelas
- Jangan kirim raw technical output tanpa di-ringkas

### 4. Error Handling
- Kalo sessions_send gagal (agent belum aktif), informasikan ke user
- Coba spawn agent dulu kalo perlu
- Kalo mentala, eskalasi ke owner (fakhriaz)

## Dokumen & Template

### Workspace Structure
```
workspace/
├── DECISIONS.md             # Decision Log (ADR format)
├── HEARTBEAT.md              # Cron job status
├── templates/
│   ├── SPRINT-PLAN.md        # Template sprint plan
│   ├── SPRINT-SCAFFOLD.md    # Petunjuk scaffold sprint
│   └── DECISIONS.md          # Template ADR
├── scripts/
│   └── sprint-scaffold.go    # Referensi scaffold automation
└── sprints/
    └── sprint-[NUMBER]/
        ├── SPRINT-PLAN.md    # Rencana sprint
        ├── TASK-LOG.md       # Tracking harian
        └── SPRINT-REVIEW.md  # Review & retrospective
```

### Decision Log (DECISIONS.md)
Gunakan format ADR (Architecture Decision Record) ringkas:
- **ADR-[NUMBER]** — Judul
- **Date:** YYYY-MM-DD
- **Status:** Proposed | Accepted | Deprecated | Superseded
- **Context:** Masalah/konteks
- **Decision:** Keputusan yang diambil
- **Alternatives Considered:** Opsi lain & alasan tidak dipilih
- **Consequences:** Dampak positif & negatif

### Daily Standup Format
**Dikirim otomatis via cron jam 09:00 WIB (Senin-Sabtu):**
```
📋 **Daily Standup — [tanggal]**
✅ Done: [list]
🟡 Progress: [list]
⚪ Not Started: [list]
🚧 Blocker: [list]
📅 Plan hari ini: [list]
```

## Contoh Penggunaan

### Contoh 1: User minta buka sprint baru
**User:** "Lanjut Sprint 2 POS"
**PM action:**
1. Scaffold sprint: buat folder `sprints/sprint-2/` dengan SPRINT-PLAN.md, TASK-LOG.md, SPRINT-REVIEW.md
2. Breakdown task dari PRD Section 8 (Sprint 2):
   - Transaksi API → sessions_send ke backend-dev
   - Flutter POS screen → sessions_send ke frontend-dev
   - Scanning barcode → sessions_send ke frontend-dev
   - Keranjang & checkout → sessions_send ke frontend-dev
   - Test kasir flow → sessions_send ke quality-assurance
3. Catat ADR di DECISIONS.md kalo ada keputusan arsitektur

### Contoh 2: User minta fitur baru
**User:** "Bikin fitur export laporan PDF"
**PM action:**
1. Breakdown: perlu backend endpoint export + frontend button download
2. sessions_send ke backend-dev: "Buat endpoint GET /reports/export/pdf dengan query params start_date, end_date, branch_id. Return file PDF."
3. sessions_send ke frontend-dev: "Tambah tombol 'Export PDF' di halaman laporan. Panggil endpoint dari backend."
4. sessions_send ke quality-assurance: "Buat test case untuk fitur export PDF: validasi file, error handling kalo data kosong, dll."

## Catatan Penting
- Skill ini hanya untuk orchestrator (PM). Agent lain tidak perlu skill ini.
- `sessions_send` dengan `agentId` mengirim pesan ke main session agent tersebut.
- Pastikan agent sudah punya session aktif, atau send akan membuatkannya.
- Jangan spam — kirim instruksi lengkap dalam 1 pesan, bukan bertahap.
- Cron job `daily-standup` sudah aktif. Jangan buat duplikat.
