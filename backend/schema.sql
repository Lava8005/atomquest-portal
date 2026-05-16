-- Core Database Schema for In-House Goal Setting & Tracking Portal
-- Humanized Comments: High modularity, explicit index constraints, no intermediate logic overhead.

CREATE TYPE user_role AS ENUM ('Employee', 'Manager', 'Admin');
CREATE TYPE goal_uom AS ENUM ('Numeric', 'Percentage', 'Timeline', 'Zero-based');
CREATE TYPE goal_status AS ENUM ('Not Started', 'On Track', 'Completed');

-- 1. Users Table (Handles org hierarchy and roles)
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    role user_role NOT NULL DEFAULT 'Employee',
    manager_id INT REFERENCES users(id) ON DELETE SET NULL,
    department VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 2. Goal Cycles Table (Enforces strict systemic tracking windows)
CREATE TABLE goal_cycles (
    id SERIAL PRIMARY KEY,
    year INT NOT NULL,
    phase1_open TIMESTAMP WITH TIME ZONE NOT NULL,
    q1_open TIMESTAMP WITH TIME ZONE NOT NULL,
    q2_open TIMESTAMP WITH TIME ZONE NOT NULL,
    q3_open TIMESTAMP WITH TIME ZONE NOT NULL,
    q4_open TIMESTAMP WITH TIME ZONE NOT NULL,
    is_locked BOOLEAN DEFAULT FALSE
);

-- 3. Goal Sheets Table (Aggregates individual cycle metrics)
CREATE TABLE goal_sheets (
    id SERIAL PRIMARY KEY,
    user_id INT REFERENCES users(id) ON DELETE CASCADE,
    cycle_id INT REFERENCES goal_cycles(id) ON DELETE RESTRICT,
    status VARCHAR(50) DEFAULT 'Draft', -- Draft, Pending Approval, Approved, Rework
    approved_by INT REFERENCES users(id),
    approved_at TIMESTAMP WITH TIME ZONE,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_user_cycle UNIQUE (user_id, cycle_id)
);

-- 4. Goals Table (Handles granular tracking fields and validation frameworks)
CREATE TABLE goals (
    id SERIAL PRIMARY KEY,
    sheet_id INT REFERENCES goal_sheets(id) ON DELETE CASCADE,
    thrust_area VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    uom goal_uom NOT NULL,
    target_value NUMERIC(12, 2) NOT NULL,
    weightage INT NOT NULL CHECK (weightage >= 10), -- Enforces min 10% weightage
    parent_goal_id INT REFERENCES goals(id) ON DELETE SET NULL, -- Self-referencing for Shared Departmental KPIs
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 5. Quarterly Progress Tracking Table
CREATE TABLE goal_progress (
    id SERIAL PRIMARY KEY,
    goal_id INT REFERENCES goals(id) ON DELETE CASCADE,
    quarter VARCHAR(2) NOT NULL, -- Q1, Q2, Q3, Q4
    actual_achievement NUMERIC(12, 2) DEFAULT 0.00,
    status goal_status NOT NULL DEFAULT 'Not Started',
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_goal_quarter UNIQUE (goal_id, quarter)
);

-- 6. Structured Manager Comments Table
CREATE TABLE manager_reviews (
    id SERIAL PRIMARY KEY,
    sheet_id INT REFERENCES goal_sheets(id) ON DELETE CASCADE,
    quarter VARCHAR(2) NOT NULL,
    reviewer_id INT REFERENCES users(id),
    comments TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT unique_sheet_quarter_review UNIQUE (sheet_id, quarter)
);

-- 7. Audit Trail Table (Captures systemic changes securely after goal lock)
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    goal_id INT NOT NULL,
    changed_by INT REFERENCES users(id),
    action_type VARCHAR(50) NOT NULL, -- UPDATE, UNLOCK, OVERRIDE
    old_value JSONB,
    new_value JSONB,
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Optimization Indices for Rapid Dashboard Generation & Analytics Processing
CREATE INDEX idx_users_manager ON users(manager_id);
CREATE INDEX idx_goals_sheet ON goals(sheet_id);
CREATE INDEX idx_progress_goal ON goal_progress(goal_id);
CREATE INDEX idx_audit_goal ON audit_logs(goal_id);