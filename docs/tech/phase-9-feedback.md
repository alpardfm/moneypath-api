# Phase 9 Feedback

This document records the first operational notes gathered from the production-like deployment and smoke test flow.

## What Felt Good

- deploy flow is reproducible through Docker
- migrations are executed automatically before API startup
- health endpoint is enough to confirm app and database readiness
- core finance flows can be checked with one smoke-test script

## Friction Notes

- deployment target is still generic; no cloud-specific manifest exists yet
- smoke test validates core success paths only, not failure scenarios
- no seed/admin bootstrap flow exists for first-time production setup
- no metrics/observability stack exists beyond structured request logs
- no backup and restore procedure is documented yet

## Improvement Backlog

- choose a concrete deployment platform and add its manifest
- add rollback notes for failed deploys
- add database backup strategy documentation
- add smoke tests for negative-balance rejection and debt edge cases
- add monitoring and alerting baseline
- add release versioning or image tagging strategy
