# Moneypath API

Moneypath API is a backend service for a single-user personal finance application designed to help users track money movement, wallet balances, debts, and basic financial condition through a simple but structured system.

This project is built not only as a backend exercise, but as a real product foundation intended for daily personal use, future frontend integration, deployment, and iterative improvement based on real-world usage.

---

## Vision

Moneypath is a personal finance product focused on:

- tracking money flow clearly
- managing multiple wallets
- recording debt and debt payments
- keeping wallet balances trustworthy
- providing basic financial summary and dashboard
- building a usable daily finance workflow before chasing advanced features

The first version prioritizes **trust, correctness, and usability** over complexity.

---

## Product Principles

### 1. Mutation is the source of truth
All balance-changing events must go through `mutation`.

This means:

- wallet balance is not manually edited
- debt financial state is not freely manipulated
- summary and dashboard are derived from actual financial events

### 2. Wallet and Debt are master data
Wallet and Debt are master records. Mutation is the event layer that changes financial state.

### 3. Trust over features
This is a finance product. If wallet balance or debt state becomes inconsistent, user trust is broken.

### 4. Real usage over theoretical completeness
The MVP is built to be actually used in daily life, not just to look complete on paper.

---

## MVP Scope

### Included
- Auth
- Profile
- Wallet management
- Mutation management
- Debt management
- Basic dashboard
- Basic summary

### Postponed
- advanced analytics
- category reporting
- recurring transactions
- notifications
- export/report
- multi-user collaboration
- advanced auth flows
- richer settings and preferences

---

## Core Domain

### Master Data
- User
- Wallet
- Debt

### Financial Event
- Mutation

### Derived Data
- Dashboard
- Summary

---

## Main Business Rules

## Wallet
- a user can create multiple wallets
- wallet balance cannot be edited directly
- all wallet balance changes must go through mutations
- wallet balance must not go below zero
- outgoing mutation must be rejected if balance is insufficient
- wallet cannot be deleted/inactivated if balance is not zero
- wallet can be soft deleted/inactivated only if balance is zero
- inactive wallet should not appear in active selection
- wallet history must remain intact

## Debt
- debt is master data
- debt can be created directly from debt module
- debt can also be created from mutation flow
- debt state is updated through related mutations
- debt cannot be deleted if remaining debt is not zero
- fully paid debt can remain with status `lunas`
- fully paid debt may also be soft deleted/inactivated
- debt history must remain intact

## Mutation
- mutation is the main source of financial events
- for MVP, mutation has only two types:
  - `masuk`
  - `keluar`
- minimum fields:
  - date
  - time
  - type
  - wallet
  - amount
  - description
  - related_to_debt (yes/no)
- `masuk` increases wallet balance
- `keluar` decreases wallet balance
- outgoing mutation must be rejected if wallet balance is insufficient
- mutation can be edited
- mutation cannot be deleted

## Mutation and Debt Interaction

If mutation is related to debt, a simple toggle must exist:

- `related_to_debt = yes/no`

### If mutation is `keluar` and related to debt
- user chooses source wallet
- user chooses debt
- wallet balance is reduced
- debt remaining amount is reduced
- mutation is recorded normally

### If mutation is `masuk` and related to debt
- user can choose existing debt
- or create new debt

### If creating new debt from mutation
- mutation amount and debt initial amount may be different
- this supports real-world cases such as admin fees, deductions, or partial disbursement

Required fields when creating debt from mutation:
- debt name
- initial debt amount
- tenor
- tenor unit
- installment/payment amount
- other required debt metadata

---

## MVP Success Criteria

Moneypath MVP is considered successful if:

- user can register and login
- user can create wallets
- user can record incoming and outgoing mutations
- wallet balances update correctly
- wallet cannot go negative
- user can create and manage debts
- mutations can interact with debts correctly
- user can see a simple dashboard and summary
- the API is deployable
- the product is usable for daily personal finance tracking

---

## Suggested Architecture

Recommended layered architecture:

- handler / controller
- service / use case
- repository
- domain / entity
- database / migration

Suggested modules:

- auth
- profile
- wallet
- debt
- mutation
- dashboard
- summary

---

## Data Ownership

All core data must belong to the authenticated user through `user_id`.

This means:

- user cannot access another user's wallets
- user cannot access another user's debts
- user cannot access another user's mutations
- all reads and writes must enforce ownership rules

---

## Critical Engineering Concerns

This project must handle carefully:

- wallet balance consistency
- mutation edit recalculation logic
- debt remaining amount consistency
- safe transactional updates
- prevention of negative balances
- prevention of cross-user data leakage

---

## QA Priorities

### Tier 1
- register/login works correctly
- user isolation works correctly
- incoming mutation updates wallet correctly
- outgoing mutation updates wallet correctly
- wallet cannot go negative
- debt payment reduces remaining debt correctly
- editing mutation does not corrupt wallet/debt state

### Tier 2
- creating debt from mutation is consistent
- wallet cannot be deleted when balance is not zero
- debt cannot be deleted when remaining debt is not zero
- summary calculations are correct

### Tier 3
- filtering mutation history
- profile update behavior
- debt status display
- inactive/archive display behavior

---

## Roadmap Overview

### Phase 0
Foundation and project setup

### Phase 1
Domain design and database schema

### Phase 2
Auth and profile

### Phase 3
Wallet module

### Phase 4
Debt module

### Phase 5
Mutation core

### Phase 6
Mutation and debt integration

### Phase 7
Summary and dashboard

### Phase 8
API hardening and testing

### Phase 9
Deployment and real-life usage

---

## Long-Term Intent

Moneypath should become a real end-to-end product that is:

- usable in daily life
- technically trustworthy
- deployable
- testable
- shareable for feedback
- built with both engineering and product thinking

---

## One-Line System Decision

Moneypath is a single-user personal finance product where wallets and debts are master data, all balance-changing events go through mutations, and the first release is focused on delivering a usable, trustworthy daily finance workflow rather than an overbuilt finance platform.