# API Hardening

This document captures the API hardening rules introduced in Phase 8.

## Response Shape

Every endpoint returns the same top-level envelope:

```json
{
  "success": true,
  "data": {},
  "meta": {},
  "error": null
}
```

Error responses use:

```json
{
  "success": false,
  "error": {
    "code": "validation_error",
    "message": "wallet name is required"
  }
}
```

## Shared Error Codes

- `invalid_json`: request body cannot be decoded
- `validation_error`: request shape or query params are invalid
- `unauthorized`: bearer token is missing or invalid
- `internal_error`: unexpected server-side failure

Module-specific error codes stay explicit, for example:

- `wallet_not_found`
- `debt_not_found`
- `mutation_not_found`
- `insufficient_wallet_balance`
- `mutation_delete_not_allowed`

## Pagination

List endpoints support:

- `page`
- `page_size`

Default behavior:

- `page=1`
- `page_size=20`
- maximum `page_size=100`

Pagination metadata is returned in `meta`:

```json
{
  "page": 1,
  "page_size": 20,
  "total_items": 42,
  "total_pages": 3
}
```

## List Endpoints

Paginated endpoints:

- `GET /wallets`
- `GET /debts`
- `GET /mutations`

## Mutation Filters

`GET /mutations` supports:

- `type=masuk|keluar`
- `wallet_id=<uuid>`
- `debt_id=<uuid>`
- `related_to_debt=true|false`
- `from=<RFC3339 timestamp>`
- `to=<RFC3339 timestamp>`
- `sort_by=happened_at|created_at|amount`
- `sort_direction=asc|desc`

Example:

```text
GET /mutations?type=keluar&related_to_debt=true&sort_by=happened_at&sort_direction=desc&page=1&page_size=10
```

## Naming Consistency

- JSON fields use `snake_case`
- path params use resource ids like `walletID`, `debtID`, `mutationID` in router code, but serialized JSON remains `snake_case`
- response envelopes always use `success`, `data`, `meta`, `error`
- mutation domain terms stay aligned with product language: `masuk`, `keluar`, `lunas`

## Logging

Request logs now include:

- `method`
- `path`
- `status`
- `request_id`
- `user_id`
- `duration`

This keeps audit visibility lightweight without coupling business modules to logging concerns.
