# AtomQuest Goal Setting & Tracking Portal
### AtomQuest Hackathon 1.0 — Submission by Lakshya Sharma

<div align="center">

🔗 **[Live Demo → atomquest-portal-eight.vercel.app](https://atomquest-portal-eight.vercel.app)**


</div>

---

## Demo Access — 3 Role Journeys

| Role | How to Login | Identity |
|------|-------------|----------|
| 👤 Employee | Click **"Arjun Kumar"** on the login screen | Senior Engineer, Engineering & Technology |
| 👔 Manager (L1) | Click **"R. Sharma"** on the login screen | Director of Technology, Engineering Management |
| 🛡️ Admin / HR | Click **"System Admin"** on the login screen | HR Audit Core, People Operations & Compliance |

> No password required. Click the role card to authenticate via the SSO demo routing layer.

---

## Tech Stack

| Layer | Technology | Hosting |
|-------|-----------|---------|
| Frontend | Vite + TypeScript + Tailwind CSS | Vercel (CDN) |
| Backend API | Go Fiber v2 (REST) | Render |
| Database | PostgreSQL + sqlx | Render |
| Auth | JWT + RBAC Middleware | `pkg/middleware/auth.go` |
| Notifications | Microsoft Teams Webhook (Async) | Goroutine |

---

## Architecture
<img width="707" height="851" alt="architecture-atom" src="https://github.com/user-attachments/assets/b266233d-3c3c-46a7-8d96-7cc893475afc" />

Three-tier architecture with Microsoft Entra ID SSO flow, atomic SQL transactions, and non-blocking async webhook delivery.

```
Microsoft Entra ID
       ↕  SSO Auth / JWT + Role Claims
Vite + TypeScript Frontend  (Vercel)
       ↕  REST POST/GET — Bearer JWT
Go Fiber REST API            (Render)
    ↙                    ↘
PostgreSQL DB          Microsoft Teams
(Atomic SQL / sqlx)    (Async JSON Webhook)
```

---

## BRD Compliance — Full Feature Map

### ✅ Phase 1 — Goal Creation & Approval (Section 2.1)

- **Employee goal sheet creation** — Thrust Area, Goal Title, Unit of Measure, Target Value, Weightage per goal
- **Validation Rule: Total weightage = exactly 100%** — enforced in `pkg/engine/validate.go` and live in `src/main.ts:validate()`; progress bar turns red if over/under, submit button disabled until valid
- **Validation Rule: Minimum weightage per goal = 10%** — enforced in `src/main.ts:addGoal()` with inline error message
- **Validation Rule: Maximum 8 goals per employee** — enforced in `addGoal()` with user-facing error
- **Manager (L1) Approval Workflow** — Manager dashboard fetches all `PENDING` sheets from PostgreSQL via `GET /api/v1/goals/pending` (`handlers/manager.go`)
- **Inline Approve / Return for Rework** — `PUT /api/v1/goals/sheet/:id/approve` accepts `{ status: "Approved" }` or `{ status: "Rework" }` — atomic SQL UPDATE in `handlers/manager.go:ManagerApproveGoalSheet()`
- **Goal locking on approval** — status column in `goal_sheets` table; frontend respects locked state
- **Shared Goals / Corporate KPI Broadcast** — Admin panel "Broadcast Departmental KPI" pushes a read-only corporate objective to employee goal sheets via `injectCorporateKPI()` in `src/main.ts`

---

### ✅ Phase 2 — Achievement Tracking & Quarterly Check-ins (Section 2.2)

- **Quarterly actuals entry** — Q1 / Q2 / Q3 / Q4 input fields per goal in Phase 2 tab
- **All 4 BRD Scoring Formulas implemented** in `src/main.ts:calcScore()`:

| UoM Type | Formula | Implementation |
|----------|---------|----------------|
| Numeric / % (Min — higher is better) | Achievement ÷ Target | `case '%': case 'Numeric': return actual / target` |
| Numeric / % (Max — lower is better) | Target ÷ Achievement | `case 'Timeline': return target / actual` |
| Timeline (date-based) | Target days ÷ Actual days | `if (actual === 0) return 0` guard included |
| Zero-based (0 = success) | 100% if 0, else 0% | `return actual === 0 ? 100 : 0` |

- **Status per goal** — PENDING / ON TRACK / EXCEEDED / LAGGING computed live from score
- **Save Progress → PostgreSQL UPSERT** — `saveTracking()` loops all goals, sends `POST /api/v1/progress/update` per goal with real DB IDs remapped from backend response. Handler uses `ON CONFLICT (goal_id, quarter) DO UPDATE` — no duplicate records
- **Manager Check-in Comments** — "Save Review Note" textarea in Manager panel calls `POST /api/v1/progress/review` → `handlers/progress.go:AddManagerReview()`, UPSERT to `manager_reviews` table

---

### ✅ Check-in Schedule (Section 2.3)

- Active quarter shown dynamically in the dashboard header (`current-cycle` element), computed from `new Date().getMonth()` on load
- Portal is structured around the 5-phase window: Goal Setting → Q1 → Q2 → Q3 → Q4/Annual

---

### ✅ User Roles & Personas (Section 3)

| Role | Capabilities Implemented |
|------|------------------------|
| **Employee** | Create goals, enter quarterly actuals, view locked approved sheet, Phase 2 tracking |
| **Manager (L1)** | Team pending approvals dashboard, inline approve/reject, check-in comment submission |
| **Admin / HR** | Analytics dashboard, escalation engine, KPI broadcast, CSV export, JSON report |

- Role switching via Sign Out button in header (clears session, returns to login)
- Dynamic identity header — name, role title, department update on every role switch
- RBAC enforced on every API route via `middleware.RequireRoles()` in `pkg/middleware/auth.go`

---

### ✅ Reporting & Governance (Section 4)

- **CSV Achievement Report** — `exportAchievementCSV()` generates and downloads `AtomQuest_Achievement_Audit_Report.csv` with columns: Goal ID, Thrust Area, Title, UoM, Target, Weightage, Latest Progress, Achievement %
- **Analytics Dashboard** — `GET /api/v1/analytics/dashboard` runs a single aggregation SQL query (`handlers/analytics.go`) with COALESCE for null safety; returns total goals, completed, on track, sheets approved, check-ins done
- **Audit / Escalation Log** — all SLA breaches recorded in `escalations` table with `user_id`, `manager_id`, `rule_triggered`, `status`, `created_at`

---

## Bonus Features (Section 5)

### ✅ 5.1 — Microsoft Entra ID Integration
- JWT-based authentication with embedded role claims (`AuthClaims` struct in `pkg/middleware/auth.go`)
- `GenerateToken(userID, role)` creates 12-hour signed tokens
- `Protect()` middleware validates every protected route — unauthorized requests receive `401`
- `RequireRoles(...)` RBAC middleware enforces role-level access per endpoint
- SSO demo routing layer simulates the Entra ID popup flow with realistic delay

### ✅ 5.2 — Microsoft Teams Integration
- Adaptive Card webhook payload fires on every goal sheet submission (`handlers/goals.go`)
- Wrapped in a **goroutine** — completely non-blocking, zero impact on API response time
- Payload includes employee name, submission timestamp, and deep-link anchor

### ✅ 5.3 — Escalation Module (Rule-Based)
- Two configurable escalation rules in `handlers/escalations.go`:
  - **Rule 1:** "L1 Approval Overdue (>3 Days)" — fires when `goal_sheets.status = 'PENDING'` and `updated_at < NOW() - INTERVAL '3 days'`
  - **Rule 2:** "Q1 Check-in Overdue" — fires when sheet is approved but no `goal_progress` record exists for Q1
- `ON CONFLICT DO NOTHING` prevents duplicate escalation logs
- Admin panel "⚡ Scan & Fetch Real-Time Logs" button triggers engine then fetches live log table from PostgreSQL
- Full escalation log rendered in UI with department, reason, and timestamp

### ✅ 5.4 — Analytics Module
- Real-time aggregation query in `handlers/analytics.go` — no N+1, single round-trip to DB
- Returns: total goals, goals completed, goals on track, goals not started, total sheets approved, total check-ins done
- Admin "Generate Company JSON Report" renders formatted JSON output live in the dashboard
- Foundation for QoQ trends — `goal_progress` table stores per-quarter actuals enabling historical queries

---

## Cost Optimisation (Section 6 — Evaluation Criterion 6)

| Decision | Benefit |
|----------|---------|
| **Vercel** for frontend | Free CDN-distributed hosting, zero cold starts, global edge |
| **Go Fiber** backend | Lowest memory footprint of any web framework; handles high concurrency on Render free tier |
| **Single aggregation SQL** in analytics | Eliminates N+1 queries — entire dashboard in one DB round-trip |
| **sqlx connection pool** | Reuses DB connections — no per-request connection overhead |
| **Async goroutine** for Teams webhook | Zero blocking on API thread — webhook failure never affects user response |
| **PostgreSQL UPSERTs** | `ON CONFLICT DO UPDATE` — no duplicate inserts, no read-before-write pattern |
| **ON CONFLICT DO NOTHING** in escalations | Idempotent rule engine — can run repeatedly without DB bloat |

---

## Project Structure

```
atomquest-portal/
├── main.go                        # Fiber app bootstrap, CORS, middleware chain
├── pkg/
│   ├── api/
│   │   ├── router.go              # Centralized route registry with RBAC injection
│   │   └── handlers/
│   │       ├── goals.go           # Phase 1: goal sheet creation, atomic TX, Teams webhook
│   │       ├── manager.go         # L1 approval workflow, pending sheets query
│   │       ├── progress.go        # Phase 2: quarterly UPSERT, manager review UPSERT
│   │       ├── analytics.go       # Aggregation SQL dashboard (Bonus 5.4)
│   │       └── escalations.go     # Rule-based SLA engine + log fetch (Bonus 5.3)
│   ├── middleware/
│   │   └── auth.go                # JWT generation, Protect(), RequireRoles() RBAC
│   ├── engine/
│   │   └── validate.go            # Business rules: weightage=100%, max 8, min 10%
│   └── db/
│       └── db.go                  # sqlx connection pool initializer
└── frontend/
    ├── index.html                 # Single-page app shell, Tailwind CSS
    └── src/
        ├── main.ts                # All UI logic, API calls, role routing, score calc
        ├── api.ts                 # Typed fetch wrapper with Bearer token injection
        └── auth.ts                # SSO simulation layer (Entra ID demo mode)
```

---

## API Endpoints

| Method | Route | Role | Description |
|--------|-------|------|-------------|
| `POST` | `/api/v1/goals/sheet` | Employee | Submit Phase 1 goal sheet (atomic TX) |
| `GET` | `/api/v1/goals/pending` | Manager, Admin | Fetch pending approval queue |
| `PUT` | `/api/v1/goals/sheet/:id/approve` | Manager, Admin | Approve or return for rework |
| `POST` | `/api/v1/progress/update` | Employee | Log quarterly actuals (UPSERT) |
| `POST` | `/api/v1/progress/review` | Manager, Admin | Submit check-in comment (UPSERT) |
| `GET` | `/api/v1/analytics/dashboard` | Admin | Organisation-wide aggregation metrics |
| `POST` | `/api/v1/escalations/trigger` | Admin | Run SLA breach detection rules |
| `GET` | `/api/v1/escalations/logs` | Admin | Fetch escalation audit trail |
| `GET` | `/health` | Public | Backend health check |

---

<div align="center">

Built for **AtomQuest Hackathon 1.0** · Go Fiber + PostgreSQL + TypeScript + Vercel + Render

</div>
