/*
Sprint Scaffold Script — Reference Implementation

Golang script buat generate sprint structure.
Fungsinya sebagai referensi — executable kalo butuh automation.

Usage: go run scripts/sprint-scaffold.go [sprint-number]

What it does:
- Buat folder sprints/sprint-[N]/
- Copy SPRINT-PLAN.md template + isi nomor sprint
- Generate TASK-LOG.md kosong
- Generate SPRINT-REVIEW.md template
- Append entry ke sprint timeline di SPRINTS.md
*/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/sprint-scaffold.go [sprint-number]")
		os.Exit(1)
	}

	sprintNum := os.Args[1]
	baseDir := filepath.Join("sprints", fmt.Sprintf("sprint-%s", sprintNum))

	// Create directory
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		fmt.Printf("Error creating directory: %v\n", err)
		os.Exit(1)
	}

	// Files to create
	files := map[string]string{
		"SPRINT-PLAN.md": sprintPlanTemplate(sprintNum),
		"TASK-LOG.md":    taskLogTemplate(sprintNum),
		"SPRINT-REVIEW.md": sprintReviewTemplate(sprintNum),
	}

	for name, content := range files {
		path := filepath.Join(baseDir, name)
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fmt.Printf("Error writing %s: %v\n", name, err)
			os.Exit(1)
		}
		fmt.Printf("✅ Created %s\n", path)
	}

	fmt.Printf("\n🎉 Sprint %s scaffolded successfully!\n", sprintNum)
	fmt.Printf("📁 %s/\n", baseDir)
	fmt.Printf("   ├── SPRINT-PLAN.md\n")
	fmt.Printf("   ├── TASK-LOG.md\n")
	fmt.Printf("   └── SPRINT-REVIEW.md\n")
}

func sprintPlanTemplate(num string) string {
	return fmt.Sprintf(`# Sprint %s Plan — POS Multi Branch

> **Periode:** [Start Date] — [End Date]
> **Goal:** [Tujuan sprint ini]

---

## Sprint Backlog

| # | Task | Assignee | Estimasi | Status |
|---|------|----------|----------|--------|
| 1 | [Task] | @agent | [X jam] | ⚪ Not Started |
| 2 | [Task] | @agent | [X jam] | ⚪ Not Started |

## Breakdown per Task

### Task 1: [Judul]
**Assignee:** @agent
**Acceptance Criteria:**
- [ ] Kriteria 1
- [ ] Kriteria 2

**Subtask:**
- [ ] Subtask 1
- [ ] Subtask 2

---

## Risks / Blocker
- [ ] [Risk description]
`, num)
}

func taskLogTemplate(num string) string {
	startDate := time.Now().Format("2006-01-02")
	return fmt.Sprintf(`# Sprint %s — Task Log

> **Sprint dimulai:** %s
> **Daily updates akan dicatat di sini oleh PM.**

---

## Day 1 — [Date]
**Progress:**
- Task 1: [status]
- Task 2: [status]

**Blockers:**
- [Blocker if any]

**Plan next:**
- [Next steps]

---

## Day 2 — [Date]
...
`, num, startDate)
}

func sprintReviewTemplate(num string) string {
	return fmt.Sprintf(`# Sprint %s Review

> **Tanggal:** [YYYY-MM-DD]
> **Demo oleh:** [Agent yang demo]

---

## Completed ✅
- [Task 1: brief result]
- [Task 2: brief result]

## Not Completed ❌
- [Task 1: alasan]
- [Task 2: alasan]

## Metrics
- Total task: [N]
- Completed: [N]
- Completion rate: [N]%

## Notes
[Anything notable from the sprint]

---

## Retrospective

### What went well
- [Point 1]

### What could be improved
- [Point 1]

### Action items
- [ ] Action 1 → @agent
`, num)
}