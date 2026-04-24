# KeyRaccoon Frontend

React + Vite + TypeScript + Tailwind CSS v3 admin dashboard untuk KeyRaccoon.

## Tech Stack

- **React 19** — UI library
- **Vite** — Build tool & dev server
- **TypeScript** — Type safety
- **Tailwind CSS v3** — Utility-first styling
- **React Router v7** — Client-side routing
- **Lucide React** — Icons

## Design System

Mengikuti [Raycast-inspired design system](../.dev/DESIGN.md):

- Dark theme dengan near-black blue background (`#07080a`)
- Inter font dengan positive letter-spacing
- Multi-layer shadow system untuk depth
- Raycast Red (`#FF6363`) sebagai accent color
- Opacity-based hover transitions
- macOS-native aesthetic
- Compact dan professional layout
- **Liquid Glass effects** (glassmorphism) untuk modern premium look

Lihat [LIQUID_GLASS.md](./LIQUID_GLASS.md) untuk dokumentasi lengkap efek glassmorphism.

## Quick Start

### 1. Install Dependencies

```bash
npm install
```

### 2. Configure Backend URL

```bash
# Copy environment template
cp .env.example .env

# Edit .env dan sesuaikan URL backend
# VITE_API_BASE_URL=http://localhost:8080
```

### 3. Start Development Server

```bash
npm run dev
```

Dev server berjalan di `http://localhost:5173` dan memproxy request `/api/*` dan `/health` ke backend.

## Configuration

Frontend menggunakan environment variables untuk konfigurasi dinamis. Lihat [CONFIG.md](./CONFIG.md) untuk detail lengkap.

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `VITE_API_BASE_URL` | Backend API URL | `http://localhost:8080` |
| `VITE_DEV_MODE` | Development mode | `true` |

### Contoh Konfigurasi

**Development:**
```env
VITE_API_BASE_URL=http://localhost:8080
VITE_DEV_MODE=true
```

**Production:**
```env
VITE_API_BASE_URL=https://api.production.com
VITE_DEV_MODE=false
```

## Development

```bash
# Install dependencies
npm install

# Start dev server (with API proxy)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview

# Lint code
npm run lint
```

## Project Structure

```
frontend/
├── public/              # Static assets
├── src/
│   ├── components/      # Shared components
│   │   └── Layout.tsx   # Dashboard layout with sidebar
│   ├── pages/           # Route pages
│   │   ├── LoginPage.tsx
│   │   ├── DashboardPage.tsx
│   │   ├── UsersPage.tsx
│   │   ├── ChannelsPage.tsx
│   │   ├── ProxiesPage.tsx
│   │   └── AnalyticsPage.tsx
│   ├── services/        # API service
│   │   └── api.ts       # Dynamic API client
│   ├── config/          # Configuration
│   │   └── index.ts     # App config with dynamic URL
│   ├── types/           # TypeScript types
│   │   └── index.ts
│   ├── App.tsx          # Root component with routing
│   ├── main.tsx         # Entry point
│   └── index.css        # Global styles & Tailwind
├── .env                 # Environment variables (gitignored)
├── .env.example         # Environment template
├── index.html
├── package.json
├── tailwind.config.js   # Tailwind theme customization
├── tsconfig.json
└── vite.config.ts       # Vite config with dynamic proxy
```

## Pages

| Route | Description |
|-------|-------------|
| `/login` | Authentication page |
| `/dashboard` | Overview stats & system health |
| `/users` | User management (CRUD) |
| `/channels` | AI provider channel management |
| `/proxies` | Proxy server management |
| `/analytics` | Usage statistics (placeholder) |

## Building for Production

### Standard Build

```bash
npm run build
```

Output akan berada di folder `dist/`.

### Build dengan Custom Backend URL

```bash
VITE_API_BASE_URL=https://api.production.com npm run build
```

### Docker Build

```bash
docker build --build-arg VITE_API_BASE_URL=https://api.example.com -t keyraccoon-frontend .
```

## API Integration

Frontend berkomunikasi dengan backend melalui REST API:

- **Base URL**: Dikonfigurasi via `VITE_API_BASE_URL`
- **Authentication**: JWT Bearer token
- **Endpoints**: `/api/*` untuk semua API calls
- **Health Check**: `/health` untuk system status

### Contoh Penggunaan API Service

```typescript
import { api } from './services/api';

// Login
const result = await api.login(email, password);
if (result?.access_token) {
  api.setToken(result.access_token);
}

// Get users
const users = await api.getUsers();

// Create channel
await api.createChannel({ 
  name: 'OpenAI Production', 
  type: 'openai',
  description: 'Production channel'
});
```

## Troubleshooting

### Backend Connection Issues

1. Pastikan file `.env` ada dan berisi `VITE_API_BASE_URL` yang benar
2. Verifikasi backend berjalan: `curl http://localhost:8080/health`
3. Cek browser console untuk CORS errors
4. Pastikan backend mengizinkan origin dari frontend

### Environment Changes Not Applied

1. Restart dev server setelah mengubah `.env`
2. Clear browser cache
3. Pastikan `.env` berada di folder `frontend/`

### Build Issues

```bash
# Clear cache dan reinstall
rm -rf node_modules .vite
npm install
```

## Deployment

Lihat [CONFIG.md](./CONFIG.md) untuk panduan deployment lengkap termasuk:
- Static hosting (Vercel, Netlify)
- Docker deployment
- Nginx configuration
- Environment variable setup

## Contributing

1. Ikuti design system guidelines di `.dev/DESIGN.md`
2. Gunakan TypeScript untuk type safety
3. Ikuti struktur kode dan naming conventions yang ada
4. Test di berbagai ukuran layar (mobile, tablet, desktop)
5. Pastikan accessibility (keyboard navigation, ARIA labels)
