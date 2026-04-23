# User API Keys Feature Documentation

## Overview

User API Keys adalah fitur untuk membuat dan mengelola API keys yang dapat digunakan oleh user untuk mengakses sistem KeyRaccoon. Setiap API key memiliki kontrol akses yang granular terhadap channels dan models, serta limit penggunaan dan expiration date.

## Features

### 1. API Key Management
- Create, Read, Update, Delete API keys
- Auto-generate secure API keys (format: `kr_` + 43 karakter)
- Assign API key ke specific user
- Set custom name untuk identifikasi

### 2. Access Control
- **Channels**: Specify channels yang boleh diakses (kosong = semua channel)
- **Models**: Specify models yang boleh digunakan (kosong = semua model)
- Flexible permission system

### 3. Usage Tracking
- **Usage Limit**: Set maksimal penggunaan (0 = unlimited)
- **Usage Count**: Auto-increment setiap kali digunakan
- **Last Used**: Timestamp terakhir kali digunakan

### 4. Expiration
- Set expiration date (optional)
- Auto-check expired status
- Visual indicator untuk expired keys

### 5. Security
- Show/hide API key dengan toggle
- Copy to clipboard functionality
- Masked display by default
- Secure random generation

## Database Schema

### user_api_keys
```sql
CREATE TABLE user_api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    name VARCHAR(255) NOT NULL,
    key TEXT NOT NULL UNIQUE,
    is_active BOOLEAN DEFAULT true,
    usage_limit BIGINT DEFAULT 0,
    usage_count BIGINT DEFAULT 0,
    expires_at TIMESTAMP,
    last_used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);
```

### user_api_key_channels (junction table)
```sql
CREATE TABLE user_api_key_channels (
    user_api_key_id BIGINT REFERENCES user_api_keys(id),
    channel_id BIGINT REFERENCES channels(id),
    PRIMARY KEY (user_api_key_id, channel_id)
);
```

### user_api_key_models
```sql
CREATE TABLE user_api_key_models (
    id BIGSERIAL PRIMARY KEY,
    user_api_key_id BIGINT REFERENCES user_api_keys(id),
    model_id BIGINT REFERENCES models(id),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_api_key_id, model_id)
);
```

## Backend API

### Endpoints

```
GET    /api/user-api-keys              - List all API keys (Admin)
GET    /api/user-api-keys/:id          - Get API key by ID
GET    /api/user-api-keys/user/:userID - Get API keys by user
POST   /api/user-api-keys              - Create new API key (Admin)
PUT    /api/user-api-keys/:id          - Update API key (Admin)
DELETE /api/user-api-keys/:id          - Delete API key (Admin)

POST   /api/user-api-keys/:id/channels           - Add channel
DELETE /api/user-api-keys/:id/channels/:channelID - Remove channel
POST   /api/user-api-keys/:id/models             - Add model
DELETE /api/user-api-keys/:id/models/:modelID    - Remove model
```

### Request Examples

**Create API Key:**
```json
POST /api/user-api-keys
{
  "user_id": 1,
  "name": "Production API Key",
  "usage_limit": 10000,
  "expires_at": "2026-12-31T23:59:59Z",
  "channel_ids": [1, 2],
  "model_ids": [1, 3, 5]
}
```

**Response:**
```json
{
  "id": 1,
  "user_id": 1,
  "name": "Production API Key",
  "key": "kr_abc123def456ghi789jkl012mno345pqr678stu901",
  "is_active": true,
  "usage_limit": 10000,
  "usage_count": 0,
  "expires_at": "2026-12-31T23:59:59Z",
  "last_used_at": null,
  "created_at": "2026-04-23T15:00:00Z",
  "updated_at": "2026-04-23T15:00:00Z",
  "user": {
    "id": 1,
    "email": "user@example.com",
    "name": "John Doe"
  },
  "channels": [
    {"id": 1, "name": "0penAI", "type": "openai"},
    {"id": 2, "name": "Anthr0pic", "type": "anthr0pic"}
  ],
  "models": [
    {"id": 1, "name": "gpt-4", "display_name": "GPT-4"},
    {"id": 3, "name": "claude-3", "display_name": "Claude 3"}
  ]
}
```

## Frontend UI

### UserAPIKeysPage Features

**1. Table View:**
- Name dengan icon Key
- User email (untuk admin)
- API key dengan show/hide toggle
- Copy to clipboard button
- Status badge (Active/Inactive/Expired/Limit Reached)
- Usage counter (current/limit)
- Expiration date
- Delete action

**2. Create Modal:**
- User selector (dropdown)
- Name input
- Usage limit (0 = unlimited)
- Expiration date picker (datetime-local)
- Channels multi-select (checkbox list)
- Models multi-select (checkbox list)
- Scrollable lists untuk banyak options

**3. Status Indicators:**
- 🟢 Active - API key aktif dan bisa digunakan
- 🔴 Inactive - API key dinonaktifkan
- 🔴 Expired - Sudah melewati expiration date
- 🟡 Limit Reached - Usage count >= usage limit

**4. Security Features:**
- API key masked by default (••••••••••••••••)
- Eye icon untuk show/hide
- Copy button untuk clipboard
- Alert confirmation saat copy

## Usage Flow

### Admin Creates API Key:
1. Admin navigates to API Keys page
2. Click "Create API Key" button
3. Fill form:
   - Select user
   - Enter name
   - Set usage limit (optional)
   - Set expiration date (optional)
   - Select allowed channels (optional)
   - Select allowed models (optional)
4. Submit form
5. API key generated automatically
6. Copy API key and share with user

### User Uses API Key:
1. User receives API key from admin
2. User includes API key in request header:
   ```
   Authorization: Bearer kr_abc123...
   ```
3. System validates:
   - API key exists
   - Is active
   - Not expired
   - Usage limit not reached
   - Has access to requested channel
   - Has access to requested model
4. If valid, increment usage count
5. Update last_used_at timestamp

## Permissions

### Admin:
- ✅ View all API keys
- ✅ Create API keys for any user
- ✅ Update API keys
- ✅ Delete API keys
- ✅ Manage channels and models

### Regular User:
- ✅ View own API keys only
- ❌ Cannot create API keys
- ❌ Cannot update API keys
- ❌ Cannot delete API keys

## Validation Rules

**Backend:**
1. Name is required
2. User must exist
3. Channels must exist (if specified)
4. Models must exist (if specified)
5. Usage limit >= 0
6. Expiration date must be in future (if specified)
7. API key must be unique

**Frontend:**
1. User selection required
2. Name required
3. Usage limit must be number >= 0
4. Expiration date format: RFC3339

## Helper Methods

**Model Methods:**
```go
func (k *UserAPIKey) IsExpired() bool
func (k *UserAPIKey) IsLimitReached() bool
func (k *UserAPIKey) CanUse() bool
func (k *UserAPIKey) IncrementUsage()
```

**Service Methods:**
```go
func GenerateAPIKey() (string, error)
func ValidateAPIKey(key string) (*UserAPIKey, error)
func IncrementUsage(id uint) error
```

## Best Practices

1. **Set Reasonable Limits**: Jangan set unlimited untuk production keys
2. **Use Expiration**: Set expiration date untuk temporary access
3. **Restrict Channels**: Limit ke channels yang diperlukan saja
4. **Restrict Models**: Limit ke models yang diperlukan saja
5. **Monitor Usage**: Check usage count regularly
6. **Rotate Keys**: Regenerate keys periodically
7. **Revoke Unused**: Delete atau disable keys yang tidak digunakan

## Security Considerations

1. API keys are stored in plain text (consider encryption for production)
2. Keys should be transmitted over HTTPS only
3. Never log API keys in plain text
4. Implement rate limiting per API key
5. Monitor for suspicious usage patterns
6. Implement IP whitelisting (future enhancement)

## Future Enhancements

1. API key encryption at rest
2. IP whitelisting
3. Rate limiting per key
4. Usage analytics per key
5. Auto-rotation feature
6. Webhook notifications for limit reached
7. Audit log for API key usage
8. Scopes/permissions per key
9. Multiple keys per user
10. Key regeneration without changing ID

## Files Created

**Backend:**
- `internal/models/user_api_key.go`
- `internal/database/migrations/create_user_api_keys_table.go`
- `internal/database/repositories/user_api_key_repository.go`
- `internal/services/user_api_key_service.go`
- `internal/handlers/user_api_key_handler.go`
- `internal/routes/user_api_key_routes.go`

**Frontend:**
- `frontend/src/pages/UserAPIKeysPage.tsx`
- Updated: `frontend/src/types/index.ts`
- Updated: `frontend/src/services/api.ts`
- Updated: `frontend/src/App.tsx`
- Updated: `frontend/src/components/Layout.tsx`

## Summary

Fitur User API Keys telah fully implemented dengan:
- ✅ Complete CRUD operations
- ✅ Granular access control (channels & models)
- ✅ Usage tracking dan limits
- ✅ Expiration management
- ✅ Secure key generation
- ✅ Admin-only management
- ✅ User-friendly UI dengan show/hide dan copy features
- ✅ Status indicators dan validation
- ✅ Multi-select untuk channels dan models
- ✅ Responsive design dengan glass effects

System siap digunakan untuk production!
