# 🎯 Objective Core: Enterprise Goal Tracking Portal

Objective Core is a high-performance, full-stack Goal Setting and Performance Tracking portal built for enterprise HR and management teams. 

## 🚀 Live Demo
* **Frontend Application:** [Insert your Vercel Link here]
* **Demo Credentials:** Simply click "Log in with Microsoft SSO" for the demo flow.

## 🏗️ System Architecture
*(Upload your diagram image to GitHub and replace this text with: `![Architecture](./architecture.png)`)*

### The Tech Stack
We bypassed standard student stacks (Node/Express) to build a highly optimized, enterprise-grade architecture:
* **Frontend:** Vanilla TypeScript + Vite + Tailwind CSS (Hosted on Vercel)
* **Backend:** Go (Fiber) for O(1) space-complexity data handling (Hosted on Render)
* **Database:** PostgreSQL with strict atomic SQL transactions (Hosted on Render)
* **Integrations:** Microsoft Entra ID (SSO Demo) & MS Teams Webhooks (via Goroutines)

## ✨ Core Features
1. **Phase 1 Validation Engine:** Strict rule enforcement (max 8 goals, min 10% weight, exact 100% total weightage) processed natively.
2. **Phase 2 Scoring System:** Automated performance calculations based on dynamic Units of Measure (Numeric, Percentage, Timeline, Zero-based).
3. **Role-Based Access Control:** Secure JWT middleware isolating Employee, Manager, and Admin environments.
4. **Asynchronous Webhooks:** Go routines fire background MS Teams notifications without blocking the UI thread.
5. **Analytics & Escalation API:** O(1) endpoints built for instant managerial dashboard aggregations.

## ⚙️ Local Development
```bash
# Clone the repository
git clone [https://github.com/Lava8005/atomquest-portal.git](https://github.com/Lava8005/atomquest-portal.git)

# Start the Go API (Port 8080)
cd backend
go run main.go

# Start the Vite Frontend (Port 5173)
cd frontend
npm install
npm run dev
