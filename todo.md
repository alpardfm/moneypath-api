# TODO — Personal Finance API

## Phase 1 — Foundation
- [x] Initialize project structure
- [x] Setup Go module
- [x] Choose router/framework
- [x] Setup configuration management
- [x] Setup PostgreSQL connection
- [x] Add migration system
- [x] Prepare base README
- [x] Create basic health check endpoint
- [x] Setup environment example file

## Phase 2 — Core Domain Design
- [ ] Define main entities and relationships
- [ ] Design database schema
- [ ] Create initial schema migration SQL files
- [ ] Decide naming conventions
- [ ] Decide API response format
- [ ] Decide error handling format
- [ ] Define transaction flow rules
- [ ] Define wallet balance update strategy

## Phase 3 — Wallets & Categories
- [ ] Implement wallet entity
- [ ] Create wallet CRUD endpoints
- [ ] Implement category entity
- [ ] Create category CRUD endpoints
- [ ] Add validation rules for wallet and category

## Phase 4 — Transactions
- [ ] Implement transaction entity
- [ ] Create income transaction endpoint
- [ ] Create expense transaction endpoint
- [ ] List transactions with filters
- [ ] Add transaction detail endpoint
- [ ] Add update/delete transaction flow
- [ ] Ensure wallet balance stays consistent
- [ ] Handle transaction category relation

## Phase 5 — Debt Tracking
- [ ] Implement debt entity
- [ ] Add create/list/detail debt endpoints
- [ ] Add debt payment tracking
- [ ] Track remaining debt balance
- [ ] Add installment progress logic
- [ ] Define status: active / paid / overdue (optional)

## Phase 6 — Monthly Summary
- [ ] Build monthly income summary
- [ ] Build monthly expense summary
- [ ] Build summary by category
- [ ] Build summary by wallet
- [ ] Build debt payment summary
- [ ] Build net cash flow summary

## Phase 7 — Budgeting
- [ ] Implement monthly budget entity
- [ ] Add budget per category
- [ ] Compare actual vs budget
- [ ] Add simple over-budget indicator
- [ ] Add summary endpoint for budget usage

## Phase 8 — Code Quality
- [ ] Refactor project structure if needed
- [ ] Improve service/repository separation
- [ ] Standardize validation
- [ ] Standardize error handling
- [ ] Add logging
- [ ] Add unit tests for critical logic
- [ ] Add seed/sample data
- [ ] Add Makefile or helper scripts

## Phase 9 — Useful Improvements
- [ ] Add pagination and filtering
- [ ] Add transaction notes
- [ ] Add recurring transaction concept
- [ ] Add dashboard endpoint
- [ ] Add export-ready summary structure
- [ ] Add API documentation
- [ ] Add Swagger / OpenAPI documentation

## Phase 10 — Polish
- [ ] Review naming consistency
- [ ] Remove dead code
- [ ] Improve README
- [ ] Review API contract
- [ ] Review migration cleanliness
- [ ] Prepare project for public portfolio
