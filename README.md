# Weather App

A sleek weather dashboard built with Go backend and React + TypeScript frontend.

![Dark Mode](./img/dark.png)
![Light Mode](./img/light.png)

## Features

- Dark mode toggle
- React Query for caching weather and favorites
- `useMemo` to optimize render performance
- PostgreSQL migration and seed scripts
- Weather lookup via Open-Meteo geocoding and forecast APIs

## Setup

### Backend

1. Copy `.env.example` to `backend/.env` and update `DATABASE_URL`.
2. From `backend`, run:

```bash
cd backend
go mod tidy
go run .
```

The backend listens on `http://localhost:8080`.

### Frontend

1. From `frontend`, install dependencies:

```bash
cd frontend
npm install
npm run dev
```

2. The app will open on `http://localhost:5173`.

## Database

The backend automatically applies migrations from `backend/migrations`.

### Migration files

- `000001_create_favorites_table.up.sql`
- `000001_create_favorites_table.down.sql`
- `000002_seed_favorites.up.sql`
- `000002_seed_favorites.down.sql`

## Environment

Create `backend/.env` with:

```env
DATABASE_URL=postgresql://user:password@host:port/dbname?sslmode=require
PORT=8080
```

Create `frontend/.env` with:

```env
VITE_API_BASE_URL=http://localhost:8080
```

## Notes

- The app stores favorite cities in PostgreSQL and uses Open-Meteo for weather data.
- Make sure your DB is reachable from the backend before starting.
