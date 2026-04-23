# Role-Based Access Control (RBAC) Documentation

## Overview

KeyRaccoon implements a comprehensive Role-Based Access Control system to manage user permissions across the application.

## User Roles

### 1. SuperAdmin
- **Full system access**
- Can manage all resources
- Can create/edit/delete users, channels, proxies
- Can view all analytics

### 2. Admin
- **Administrative access**
- Can manage channels, proxies, and users
- Can view all analytics
- Cannot modify superadmin accounts

### 3. User
- **Limited access**
- Can view channels (read-only)
- Can view analytics for their bound channels only
- Cannot create, edit, or delete any resources

## Backend Implementation

### Middleware

Located in `internal/middleware/auth.go`:

```go
// Require authentication
middleware.AuthMiddleware

// Require admin or superadmin
middleware.AdminMiddleware

// Require superadmin only
middleware.SuperAdminMiddleware

// Custom role check
middleware.RoleMiddleware("admin", "superadmin")
```

### Route Protection

#### User Routes (`internal/routes/user_routes.go`)
```go
users := router.Group("/users", middleware.AuthMiddleware)
users.Post("", middleware.AdminMiddleware, userHandler.CreateUser)      // Admin only
users.Get("", middleware.AdminMiddleware, userHandler.GetAllUsers)      // Admin only
users.Get("/:id", userHandler.GetUser)                                  // All authenticated
users.Put("/:id", middleware.AdminMiddleware, userHandler.UpdateUser)   // Admin only
users.Delete("/:id", middleware.AdminMiddleware, userHandler.DeleteUser) // Admin only
```

#### Channel Routes (`internal/routes/channel_routes.go`)
```go
channels := router.Group("/channels", middleware.AuthMiddleware)
channels.Get("", channelHandler.GetAllChannels)                         // All authenticated (view-only for users)
channels.Get("/:id", channelHandler.GetChannel)                         // All authenticated
channels.Post("", middleware.AdminMiddleware, channelHandler.CreateChannel)     // Admin only
channels.Put("/:id", middleware.AdminMiddleware, channelHandler.UpdateChannel)  // Admin only
channels.Delete("/:id", middleware.AdminMiddleware, channelHandler.DeleteChannel) // Admin only
```

#### Proxy Routes (`internal/routes/proxy_routes.go`)
```go
proxies := router.Group("/proxies", middleware.AuthMiddleware, middleware.AdminMiddleware)
// All proxy operations require admin role
```

### JWT Token Structure

JWT tokens include role information:

```json
{
  "user_id": 1,
  "email": "admin@example.com",
  "role": "admin",
  "token_type": "access",
  "exp": 1234567890,
  "iat": 1234567890,
  "nbf": 1234567890
}
```

## Frontend Implementation

### Auth Context

Located in `frontend/src/contexts/AuthContext.tsx`:

```typescript
interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isAdmin: boolean;
  isSuperAdmin: boolean;
  login: (email: string, password: string) => Promise<boolean>;
  logout: () => void;
  hasPermission: (permission: string) => boolean;
}
```

### Permission System

Permissions are checked using `hasPermission()`:

```typescript
// Permission strings
'view:dashboard'    // All authenticated users
'view:channels'     // All authenticated users
'edit:channels'     // Admin and SuperAdmin only
'delete:channels'   // Admin and SuperAdmin only
'view:users'        // Admin and SuperAdmin only
'edit:users'        // Admin and SuperAdmin only
'view:proxies'      // Admin and SuperAdmin only
'edit:proxies'      // Admin and SuperAdmin only
'view:analytics'    // All authenticated users
```

### Conditional Rendering

#### Sidebar Menu

```typescript
// Layout.tsx
const navItems = [
  { path: '/dashboard', label: 'Dashboard', icon: LayoutDashboard, permission: 'view:dashboard' },
  { path: '/users', label: 'Users', icon: Users, permission: 'view:users' },
  { path: '/channels', label: 'Channels', icon: Link2, permission: 'view:channels' },
  { path: '/proxies', label: 'Proxies', icon: Shield, permission: 'view:proxies' },
  { path: '/analytics', label: 'Analytics', icon: BarChart3, permission: 'view:analytics' },
];

// Filter based on permissions
const visibleNavItems = navItems.filter(item => hasPermission(item.permission));
```

#### View-Only Mode

```typescript
// ChannelsPage.tsx
const { hasPermission } = useAuth();
const canEdit = hasPermission('edit:channels');
const canDelete = hasPermission('delete:channels');

// Conditional button rendering
{canEdit && (
  <button onClick={() => setModalOpen(true)}>
    Add Channel
  </button>
)}

// Conditional actions column
{(canEdit || canDelete) && (
  <td>
    {canEdit && <button>Edit</button>}
    {canDelete && <button>Delete</button>}
  </td>
)}
```

## Access Matrix

| Resource | SuperAdmin | Admin | User |
|----------|-----------|-------|------|
| **Dashboard** |
| View | ✅ | ✅ | ✅ |
| **Users** |
| View List | ✅ | ✅ | ❌ |
| Create | ✅ | ✅ | ❌ |
| Edit | ✅ | ✅ | ❌ |
| Delete | ✅ | ✅ | ❌ |
| **Channels** |
| View List | ✅ | ✅ | ✅ (view-only) |
| View Details | ✅ | ✅ | ✅ (view-only) |
| Create | ✅ | ✅ | ❌ |
| Edit | ✅ | ✅ | ❌ |
| Delete | ✅ | ✅ | ❌ |
| Manage API Keys | ✅ | ✅ | ❌ |
| Manage Models | ✅ | ✅ | ❌ |
| Bind Users | ✅ | ✅ | ❌ |
| **Proxies** |
| View List | ✅ | ✅ | ❌ |
| Create | ✅ | ✅ | ❌ |
| Edit | ✅ | ✅ | ❌ |
| Delete | ✅ | ✅ | ❌ |
| **Analytics** |
| View All | ✅ | ✅ | ❌ |
| View Own Channels | ✅ | ✅ | ✅ |

## Usage Examples

### Backend - Protecting Routes

```go
// Require authentication only
router.Get("/profile", middleware.AuthMiddleware, handler.GetProfile)

// Require admin role
router.Post("/channels", middleware.AuthMiddleware, middleware.AdminMiddleware, handler.CreateChannel)

// Require superadmin role
router.Delete("/users/:id", middleware.AuthMiddleware, middleware.SuperAdminMiddleware, handler.DeleteUser)

// Custom role check
router.Get("/reports", middleware.AuthMiddleware, middleware.RoleMiddleware("admin", "superadmin"), handler.GetReports)
```

### Backend - Checking Roles in Handlers

```go
func (h *Handler) GetChannels(c *fiber.Ctx) error {
    userRole, _ := c.Locals("user_role").(string)
    userID, _ := c.Locals("user_id").(uint)
    
    if userRole == "user" {
        // Return only channels bound to this user
        return h.service.GetUserChannels(userID)
    }
    
    // Return all channels for admin/superadmin
    return h.service.GetAllChannels()
}
```

### Frontend - Conditional Rendering

```typescript
import { useAuth } from '../contexts/AuthContext';

function MyComponent() {
  const { hasPermission, isAdmin } = useAuth();
  
  return (
    <div>
      {/* Show to all authenticated users */}
      <Dashboard />
      
      {/* Show only to admins */}
      {isAdmin && <AdminPanel />}
      
      {/* Check specific permission */}
      {hasPermission('edit:channels') && (
        <button>Edit Channel</button>
      )}
      
      {/* View-only mode */}
      {hasPermission('view:channels') && !hasPermission('edit:channels') && (
        <div>View Only Mode</div>
      )}
    </div>
  );
}
```

### Frontend - Protected Routes

```typescript
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

function AdminRoute({ children }: { children: React.ReactNode }) {
  const { isAdmin } = useAuth();
  
  if (!isAdmin) {
    return <Navigate to="/dashboard" replace />;
  }
  
  return <>{children}</>;
}

// Usage in App.tsx
<Route path="/users" element={
  <AdminRoute>
    <UsersPage />
  </AdminRoute>
} />
```

## Security Best Practices

1. **Always validate on backend**: Frontend checks are for UX only, not security
2. **Use middleware consistently**: Apply auth middleware to all protected routes
3. **Check permissions in handlers**: Verify user permissions before executing operations
4. **Audit role changes**: Log when user roles are modified
5. **Principle of least privilege**: Give users minimum required permissions
6. **Regular review**: Periodically review and update permission matrix

## Testing RBAC

### Backend Tests

```go
func TestAdminMiddleware(t *testing.T) {
    // Test with admin user
    // Test with regular user
    // Test with no auth
}
```

### Frontend Tests

```typescript
describe('ChannelsPage', () => {
  it('shows edit button for admin', () => {
    // Mock admin user
    // Render component
    // Assert edit button is visible
  });
  
  it('hides edit button for regular user', () => {
    // Mock regular user
    // Render component
    // Assert edit button is not visible
  });
});
```

## Troubleshooting

### User can't see menu items
- Check JWT token contains correct role
- Verify `hasPermission()` logic in AuthContext
- Check browser console for errors

### API returns 403 Forbidden
- Verify user has correct role in database
- Check JWT token is valid and contains role
- Verify middleware is applied to route
- Check backend logs for auth errors

### Role not updating after change
- User needs to logout and login again
- JWT token contains old role until refresh
- Consider implementing token refresh mechanism

## Future Enhancements

1. **Fine-grained permissions**: Move from role-based to permission-based
2. **Dynamic permissions**: Load permissions from database
3. **Permission inheritance**: Roles can inherit from other roles
4. **Resource-level permissions**: Per-channel or per-proxy permissions
5. **Audit logging**: Track all permission checks and access attempts
6. **Permission UI**: Admin interface to manage roles and permissions
