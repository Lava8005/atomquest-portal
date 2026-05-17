# AtomQuest Goal Setting & Tracking Portal
### AtomQuest Hackathon 1.0 Submission

🔗 Live Demo: https://atomquest-portal-eight.vercel.app  
📁 Stack: TypeScript + Vite (Vercel) · Go Fiber REST API (Render) · PostgreSQL

## Demo Credentials
| Role | How to Access |
|------|--------------|
| Employee | Click "Arjun Kumar" on login screen |
| Manager | Click "R. Sharma" on login screen |
| Admin/HR | Click "System Admin" on login screen |

## BRD Compliance

### Phase 1 — Goal Creation & Approval ✅
- Employee goal sheet creation with Thrust Area, Title, UoM, Target, Weightage
- Validation enforced: total weightage must equal exactly 100%
- Minimum weightage per goal: 10% | Maximum goals per employee: 8
- All rules enforced in `pkg/engine/validate.go` and `src/main.ts`
- Manager L1 approval workflow: review, approve, or return for rework
- Goals lock on approval — no further edits without Admin intervention
- Shared Goals: Admin broadcasts corporate KPI to employee sheets (read-only title/target)

### Phase 2 — Achievement Tracking & Quarterly Check-ins ✅
- Quarterly actuals entry for Q1/Q2/Q3/Q4 per goal
- All 4 BRD scoring formulas implemented:
  - Numeric/% (Min): Achievement ÷ Target
  - Numeric/% (Max): Target ÷ Achievement  
  - Timeline: Target days ÷ Actual days
  - Zero-based: 100% if 0, else 0%
- Manager check-in comments via structured review module
- Status per goal: Not Started / On Track / Completed / Exceeded

### Check-in Schedule ✅
Portal enforces quarterly windows. Active period shown dynamically on dashboard.

### User Roles ✅
All 3 roles with differentiated access:
- **Employee**: Create goals, enter actuals, view locked sheet
- **Manager (L1)**: Approve/reject goals, add check-in comments, team dashboard
- **Admin/HR**: Analytics, escalation engine, cycle management, KPI broadcast

### Reporting & Governance ✅
- CSV Achievement Report export: `exportAchievementCSV()` in `src/main.ts`
- Analytics dashboard with real-time DB aggregation
- Escalation log: audit trail of all violations with timestamps

## Bonus Features Implemented

### 5.1 Microsoft Entra ID ✅
JWT-based authentication with role claims. RBAC middleware chain in
`pkg/middleware/auth.go`. Token validation on every protected route.

### 5.2 Microsoft Teams Integration ✅
Async webhook goroutine fires on goal submission (`handlers/goals.go`).
Non-blocking — zero impact on API response time.

### 5.3 Escalation Module ✅
Rule-based engine in `handlers/escalations.go`:
- Rule 1: L1 Approval Overdue (>3 days)
- Rule 2: Q1 Check-in Overdue
Uses `ON CONFLICT DO NOTHING` — no duplicate logs.

### 5.4 Analytics Module ✅
Single aggregation SQL query with COALESCE in `handlers/analytics.go`.
Returns: goal distribution, completion rates, manager effectiveness metrics.

## Architecture
<img width="707" height="851" alt="architecture-atom" src="https://github.com/user-attachments/assets/4bdad6f7-18d2-4d33-88bc-41e1bd5a5296" />


## Cost Optimisation
- Frontend: Vercel free tier (CDN-distributed, zero cold starts)
- Backend: Render free tier (Go Fiber — minimal memory footprint)
- Database: PostgreSQL connection pooling via sqlx
- No N+1 queries — all analytics in single aggregation SQL
- Async Teams webhook — non-blocking goroutine
