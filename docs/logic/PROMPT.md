# Master Prompt — Moneypath API Vibe Coding

Use this prompt when working with an AI coding agent to keep the implementation focused, modular, and aligned with the product rules.

---

You are helping me build a backend project called **moneypath-api**.

## Product Context

Moneypath is a **single-user personal finance product**.  
Its goal is to help users track:

- wallet balances
- money movement
- debt and debt payments
- simple financial summaries

This is not an overbuilt finance platform.  
The MVP must prioritize:

- correctness
- trust
- simplicity
- usability
- clean architecture
- production-friendly code

---

## Core Domain Rules

### Entities
Core entities are:

- User
- Wallet
- Debt
- Mutation

Derived modules:

- Dashboard
- Summary

### Ownership
All data belongs to the authenticated user via `user_id`.

### Wallet Rules
- wallet is master data
- user can have multiple wallets
- wallet balance is stored but cannot be manually edited
- all wallet balance changes must go through mutations
- wallet balance must never go below zero
- wallet cannot be deleted/inactivated if balance is not zero
- inactive wallet should not appear in active selection

### Debt Rules
- debt is master data
- debt can be created directly from debt module
- debt can also be created from mutation flow
- debt has initial amount and remaining amount
- debt becomes `lunas` when remaining amount reaches zero
- debt cannot be deleted if remaining amount is not zero
- debt history must remain intact

### Mutation Rules
- mutation is the source of all balance-changing events
- mutation only has two types in MVP:
  - `masuk`
  - `keluar`
- mutation can be edited
- mutation cannot be deleted
- incoming mutation increases wallet balance
- outgoing mutation decreases wallet balance
- outgoing mutation must be rejected if wallet balance is insufficient

### Mutation + Debt Rules
- mutation has a toggle `related_to_debt`
- outgoing mutation related to debt means debt payment
- it reduces wallet balance and reduces debt remaining amount
- incoming mutation related to debt may reference existing debt or create a new debt
- mutation amount and debt initial amount may be different when creating debt from mutation
- updates affecting wallet and debt must be atomic

### Summary Rules
Summary and dashboard are derived data, not source of truth.

---

## Engineering Principles

Follow these principles strictly:

- use clean layered architecture
- avoid overengineering
- avoid unnecessary abstraction
- keep code modular and readable
- prefer explicit code over clever code
- prioritize business rule correctness
- protect data consistency
- use database transactions for critical write flows
- all queries and writes must be scoped by authenticated user
- code should be production-friendly, not toy-level

---

## Expected Architecture

Please follow this structure unless there is a strong reason not to:

- handler/controller
- request/response DTO
- service/usecase
- repository interface
- repository implementation
- domain/entity
- migration/schema
- tests

---

## How You Should Work

When I ask you to implement a module, follow this workflow:

1. First explain:
   - what files will be created
   - what files will be modified
   - what the implementation approach is

2. Then generate code for:
   - routes
   - handlers
   - DTOs
   - services
   - repositories
   - tests

3. Keep each task focused only on the requested module.
Do not expand scope unless necessary.

4. If a business rule is ambiguous, choose the simplest production-friendly option and state the assumption clearly.

5. For finance-critical logic, include test scenarios.

---

## Output Style

When generating code:
- keep it clean and minimal
- do not generate fake complexity
- do not create generic utility layers unless needed
- do not introduce event sourcing, CQRS, or microservices unless explicitly asked
- use practical naming
- make code easy to review

When generating explanations:
- be concise
- be technical
- focus on tradeoffs and correctness

---

## Important Constraints

Never break these constraints:

- wallet balance must not become negative
- mutation delete is not allowed
- edit mutation must safely reverse old effect and apply new effect
- debt remaining amount must stay consistent
- user must not access other user's data

---

## Preferred Working Mode

For each task:
- keep scope narrow
- finish one module at a time
- ensure test coverage for critical rules
- preserve consistency with previous modules

---

## Instruction Template Per Task

When I give you a task, respond in this pattern:

### 1. Plan
- files to create
- files to modify
- assumptions

### 2. Implementation
- code by layer

### 3. Tests
- important test cases
- edge cases

### 4. Notes
- potential improvements later
- risks if any

---

## Example Task Prompt

Example:

"Build the wallet module for moneypath-api using the existing architecture.  
Scope only:
- create wallet
- list active wallets
- wallet detail
- update wallet metadata
- soft delete/inactivate wallet if balance is zero

Apply all business rules from the project context.  
Also generate tests for critical validation rules."

---

## Final Reminder

The goal is not to generate the most sophisticated backend possible.  
The goal is to build a **trustworthy, usable, maintainable MVP** for a real personal finance product.