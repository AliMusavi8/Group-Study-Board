# Group Study Board

A real-time collaborative whiteboard for studying with friends. Built with Go (Gin + WebSockets), MongoDB, and Angular with Tailwind CSS.

## Project Structure

- `backend/` Go API + WebSocket server
- `frontend/` Angular client

## Prerequisites

- Go 1.25.3+
- Node.js 24.13.1 LTS+
- MongoDB 8.0+ (latest self-managed LTS; 8.1 is Atlas-only rapid release)

## Backend Setup

1. Start MongoDB locally (default `mongodb://localhost:27017`).
2. Configure environment variables:

```bash
cp backend/.env.example backend/.env
```

3. Run the server:

```bash
cd backend
# go mod tidy
# go run ./cmd/server
```

## Docker Setup

```bash
docker compose up --build
```

Services:
- Frontend: http://localhost:4200
- Backend: http://localhost:8080
- MongoDB: mongodb://localhost:27017

## Frontend Setup

```bash
cd frontend
npm install
npm run start
```

The app will be available at `http://localhost:4200`.

## Environment Variables (Backend)

- `PORT` (default `8080`)
- `MONGODB_URI` (default `mongodb://localhost:27017`)
- `DATABASE_NAME` (default `group_study_board`)
- `CORS_ORIGIN` (default `http://localhost:4200`)
- `ROOM_TTL_MINUTES` (default `60`)
- `MAX_PARTICIPANTS` (default `24`)
- `RATE_LIMIT_PER_SEC` (default `60`)
- `RATE_LIMIT_BURST` (default `120`)
- `SNAPSHOT_EVERY` (default `200`)

## Notes

- The server is authoritative for event ordering.
- Snapshots are periodically created in MongoDB to speed up late-join loads.
- Chat, shapes, and undo/redo are intentionally excluded from v1.
