# Moneypath API — Tasks by Phase

This document contains phased implementation tasks for building `moneypath-api` in a clean, realistic, MVP-first manner.

---

# Phase 0 — Foundation & Project Setup

## Goal
Prepare a clean backend foundation before implementing business logic.

## Tasks
- [x] Initialize repository `moneypath-api`
- [x] Setup folder structure with layered architecture
- [x] Add config loader for environment variables
- [x] Setup database connection
- [x] Setup migration tool
- [x] Create health check endpoint
- [x] Create standard success response format
- [x] Create standard error response format
- [x] Setup structured logger
- [x] Setup graceful shutdown
- [x] Add Dockerfile
- [x] Add docker-compose for local development
- [x] Add `.env.example`
- [x] Add linter and formatter
- [x] Add Makefile or task runner
- [x] Write initial README

## Exit Criteria
- app can run locally
- app can connect to database
- migrations can be executed
- health check endpoint works

---

# Phase 1 — Domain Design & Database Schema

## Goal
Define the core data model and entity relationships.

## Tasks
- [x] Design ERD
- [x] Create `users` table
- [x] Create `wallets` table
- [x] Create `debts` table
- [x] Create `mutations` table
- [x] Add timestamps to all core tables
- [x] Add soft delete support where needed
- [x] Define wallet active/inactive logic
- [x] Define debt active/lunas/inactive logic
- [x] Define mutation relation to wallet
- [x] Define optional mutation relation to debt
- [x] Add indexes for user-scoped queries
- [x] Create initial migrations
- [x] Document important constraints and rules

## Exit Criteria
- migrations are stable
- all core tables exist
- ownership and relations are clear
- schema supports MVP rules

---

# Phase 2 — Auth & Profile

## Goal
Build authentication and user ownership foundation.

## Tasks
- [x] Implement register endpoint
- [x] Implement login endpoint
- [x] Implement password hashing
- [x] Implement JWT/token generation
- [x] Implement auth middleware
- [x] Implement `GET /me`
- [x] Implement `PUT /me`
- [x] Implement change password endpoint
- [x] Validate duplicate email/username
- [x] Add auth unit tests
- [x] Add auth integration tests

## Exit Criteria
- user can register
- user can login
- authenticated routes are protected
- profile endpoints work

---

# Phase 3 — Wallet Module

## Goal
Build wallet management with strict balance rules.

## Tasks
- [x] Implement create wallet
- [x] Implement list active wallets
- [x] Implement wallet detail
- [x] Implement update wallet metadata
- [x] Implement soft delete / inactive wallet
- [x] Validate wallet cannot be deleted if balance != 0
- [x] Exclude inactive wallet from active selections
- [x] Validate ownership by user_id
- [x] Add wallet tests

## Exit Criteria
- user can manage wallets
- inactive wallet behavior works
- zero-balance deletion/inactivation rule works

---

# Phase 4 — Debt Module

## Goal
Build debt management as master data.

## Tasks
- [x] Implement create debt from debt menu
- [x] Implement list debts
- [x] Implement debt detail
- [x] Implement update debt metadata
- [x] Store initial amount
- [x] Store remaining amount
- [x] Derive/update `lunas` status
- [x] Validate debt cannot be deleted if remaining != 0
- [x] Allow soft delete only for paid debt
- [x] Validate ownership by user_id
- [x] Add debt tests

## Exit Criteria
- debt can be created and viewed
- remaining debt is tracked
- debt deletion rules work

---

# Phase 5 — Mutation Core

## Goal
Build the main financial event engine.

## Tasks
- [x] Implement create incoming mutation
- [x] Implement create outgoing mutation
- [x] Implement mutation history list
- [x] Implement mutation detail
- [x] Implement edit mutation
- [x] Reject outgoing mutation if balance is insufficient
- [x] Increase wallet balance for incoming mutation
- [x] Decrease wallet balance for outgoing mutation
- [x] Implement rollback-and-reapply logic for mutation edit
- [x] Validate mutation payload fields
- [x] Prevent mutation delete
- [x] Add mutation unit tests
- [x] Add mutation integration tests

## Exit Criteria
- wallet balance changes correctly from mutations
- insufficient balance is rejected
- editing mutation keeps wallet state consistent

---

# Phase 6 — Mutation + Debt Integration

## Goal
Connect debt flows with mutation flows.

## Tasks
- [x] Add `related_to_debt` toggle in mutation flow
- [x] Allow outgoing mutation to pay existing debt
- [x] Reduce debt remaining amount when paying debt
- [x] Allow incoming mutation to reference existing debt
- [x] Allow incoming mutation to create new debt
- [x] Allow mutation amount and debt initial amount to differ
- [x] Validate required fields when creating debt from mutation
- [x] Ensure wallet and debt updates happen atomically
- [x] Handle edit mutation that affects debt relation
- [x] Add integration tests for wallet-debt-mutation flows

## Exit Criteria
- debt payment through mutation works
- new debt from mutation works
- wallet and debt remain consistent

---

# Phase 7 — Summary & Dashboard

## Goal
Provide simple derived financial overview.

## Tasks
- [x] Implement total assets calculation
- [x] Implement total debts calculation
- [x] Implement balance per wallet
- [x] Implement total incoming calculation
- [x] Implement total outgoing calculation
- [x] Implement simple net flow
- [x] Create dashboard endpoint
- [x] Create summary endpoint
- [x] Add tests for calculation correctness

## Exit Criteria
- dashboard returns useful overview
- summary numbers match source data
- aggregation logic is correct

---

# Phase 8 — API Hardening

## Goal
Improve API quality, consistency, and maintainability.

## Tasks
- [x] Standardize request validation
- [x] Standardize error codes/messages
- [x] Add pagination for list endpoints
- [x] Add filter for mutation history
- [x] Add sort options for mutation history
- [x] Improve logs for critical flows
- [x] Add API documentation
- [x] Add Postman/Bruno collection
- [x] Review naming consistency
- [x] Review response consistency
- [x] Add more integration tests
- [x] Audit transaction safety

## Exit Criteria
- list endpoints are usable
- API responses are consistent
- documentation exists
- important flows are well tested

---

# Phase 9 — Deployment & Real Usage

## Goal
Deploy the product and start using it in real life.

## Tasks
- [x] Prepare production env config
- [x] Prepare production database
- [x] Add migration execution to deploy flow
- [x] Add Swagger / OpenAPI documentation
- [x] Deploy API
- [x] Verify health endpoint in deployed environment
- [x] Verify auth flow in deployed environment
- [x] Verify wallet flow in deployed environment
- [x] Verify debt flow in deployed environment
- [x] Verify mutation flow in deployed environment
- [x] Verify summary/dashboard in deployed environment
- [ ] Start personal daily usage
- [x] Record friction notes and improvement backlog

## Exit Criteria
- API is deployed
- API is stable enough for personal use
- real-world feedback starts shaping next iteration

---

# Post-MVP Backlog

## Nice to Have Later
- [ ] category system
- [ ] recurring transactions
- [ ] monthly analytics
- [ ] financial health scoring
- [ ] leakage detection
- [ ] export/report
- [ ] notifications
- [ ] richer settings
- [ ] improved archive views
- [ ] better dashboard visualizations
