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

## Development

```bash
# Install dependencies
npm install

# Start dev server (with API proxy to localhost:8080)
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

Dev server berjalan di `http://localhost:5173` dan memproxy request `/api/*` ke Go backend di `http://localhost:8080`.

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
│   │   └── api.ts
│   ├── types/           # TypeScript types
│   │   └── index.ts
│   ├── App.tsx          # Root component with routing
│   ├── main.tsx         # Entry point
│   └── index.css        # Global styles & Tailwind
├── index.html
├── package.json
├── tailwind.config.js   # Tailwind theme customization
├── tsconfig.json
└── vite.config.ts       # Vite config with API proxy
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
