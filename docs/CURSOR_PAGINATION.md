# Cursor Pagination Implementation

This wallet service now supports cursor-based pagination for transaction history, which provides better performance and consistency for large datasets compared to offset-based pagination.

## API Endpoint

```
GET /wallets/me/transactions
```

## Query Parameters

| Parameter   | Type     | Default | Description                                      |
|------------|----------|---------|--------------------------------------------------|
| `cursor`   | string   | -       | Base64-encoded cursor for pagination            |
| `limit`    | integer  | 20      | Number of transactions per page (max: 100)      |
| `direction`| string   | "next"  | Direction of pagination ("next" or "prev")      |

## Usage Examples

### First Page
```bash
GET /wallets/me/transactions?limit=10
```

### Next Page
```bash
GET /wallets/me/transactions?cursor=eyJpZCI6MTAwLCJjcmVhdGVkX2F0IjoiMjAyMy0wMS0wMVQwMDowMDowMFoifQ==&limit=10&direction=next
```

### Previous Page
```bash
GET /wallets/me/transactions?cursor=eyJpZCI6ODAsImNyZWF0ZWRfYXQiOiIyMDIzLTAxLTAxVDAwOjAwOjAwWiJ9&limit=10&direction=prev
```

## Response Format

```json
{
  "success": true,
  "message": "Transaction history retrieved successfully",
  "data": {
    "transactions": [
      {
        "id": 123,
        "created_at": "2023-01-01T00:00:00Z",
        "reference": "TRX123456",
        "wallet_id": 1,
        "transaction_type": "CREDIT",
        "amount": "100.50",
        "balance_before": "900.00",
        "balance_after": "1000.50",
        "description": "Deposit from bank",
        "status": "COMPLETED"
      }
    ],
    "pagination": {
      "page_size": 10,
      "next_cursor": "eyJpZCI6MTIzLCJjcmVhdGVkX2F0IjoiMjAyMy0wMS0wMVQwMDowMDowMFoifQ==",
      "prev_cursor": "eyJpZCI6MTMwLCJjcmVhdGVkX2F0IjoiMjAyMy0wMS0wMVQwMDowMDowMFoifQ==",
      "has_next_page": true,
      "has_prev_page": false
    }
  }
}
```

## Cursor Format

The cursor is a base64-encoded JSON object containing:
```json
{
  "id": 123,
  "created_at": "2023-01-01T00:00:00Z"
}
```

## Benefits of Cursor Pagination

1. **Consistent Results**: No duplicate or missing records when data changes during pagination
2. **Better Performance**: Doesn't require counting total records or using OFFSET
3. **Scalable**: Performance doesn't degrade with large datasets
4. **Real-time Friendly**: Handles insertions/deletions gracefully

## Implementation Details

- Transactions are ordered by `created_at DESC, id DESC` for consistent ordering
- Uses composite cursor (timestamp + ID) to handle duplicate timestamps
- Supports bidirectional navigation with `direction` parameter
- Fetches one extra record to determine if more pages exist
- Handles edge cases like empty results and invalid cursors

## Migration from Offset Pagination

The old offset-based pagination is still available for backward compatibility:

```bash
GET /wallets/me/transactions?page=1&limit=20
```

However, it's recommended to migrate to cursor pagination for better performance and consistency.
