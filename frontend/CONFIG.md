# Frontend Configuration

## Environment Variables

The frontend uses environment variables to configure the backend API URL dynamically.

### Setup

1. Copy the example environment file:
   ```bash
   cd frontend
   cp .env.example .env
   ```

2. Edit `.env` and set your backend URL:
   ```env
   VITE_API_BASE_URL=http://localhost:8080
   VITE_DEV_MODE=true
   ```

### Configuration Options

| Variable | Description | Default | Example |
|----------|-------------|---------|---------|
| `VITE_API_BASE_URL` | Backend API base URL | `http://localhost:8080` | `https://api.example.com` |
| `VITE_DEV_MODE` | Enable development mode | `true` | `false` |

### How It Works

The frontend automatically detects the backend URL in the following order:

1. **Environment Variable**: Uses `VITE_API_BASE_URL` from `.env` file
2. **Development Proxy**: In dev mode, uses Vite proxy (relative paths)
3. **Auto-detection**: Detects from `window.location` in production
4. **Fallback**: Uses `http://localhost:8080` as last resort

### Development

When running in development mode (`npm run dev`):
- Vite proxy forwards `/api` and `/health` requests to the backend
- Backend URL is configured in `vite.config.ts`
- Hot reload works automatically

```bash
npm run dev
```

### Production Build

For production builds:

```bash
# Build with default settings
npm run build

# Build with custom backend URL
VITE_API_BASE_URL=https://api.production.com npm run build
```

### Docker Deployment

When deploying with Docker, you can override the backend URL:

```bash
# Build with production API URL
docker build --build-arg VITE_API_BASE_URL=https://api.example.com -t keyraccoon-frontend .
```

Or use environment variables at runtime:

```bash
docker run -e VITE_API_BASE_URL=https://api.example.com keyraccoon-frontend
```

### Examples

#### Local Development
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_DEV_MODE=true
```

#### Staging Environment
```env
VITE_API_BASE_URL=https://staging-api.example.com
VITE_DEV_MODE=false
```

#### Production Environment
```env
VITE_API_BASE_URL=https://api.example.com
VITE_DEV_MODE=false
```

### Troubleshooting

**Issue**: Frontend can't connect to backend

**Solutions**:
1. Check `.env` file exists and has correct `VITE_API_BASE_URL`
2. Verify backend is running on the specified URL
3. Check browser console for CORS errors
4. Ensure backend allows requests from frontend origin

**Issue**: Changes to `.env` not taking effect

**Solutions**:
1. Restart the dev server (`npm run dev`)
2. Clear browser cache
3. Check if `.env` file is in the correct location (`frontend/.env`)

### Security Notes

- Never commit `.env` files to version control
- Use `.env.example` as a template
- Different environments should have different `.env` files
- Sensitive data should be in backend, not frontend environment variables
