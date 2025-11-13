# Quick Start Guide

## Prerequisites

Before starting, ensure you have:
- Go 1.21+ installed
- Node.js 18+ and npm installed
- PostgreSQL 12+ installed and running
- A PostgreSQL database named `seap` created

## Step 1: Database Setup

Create the PostgreSQL database:
```bash
createdb seap
```

Or using psql:
```bash
psql -U postgres
CREATE DATABASE seap;
\q
```

## Step 2: Backend Setup

1. Navigate to backend directory:
```bash
cd backend
```

2. Copy environment file:
```bash
cp .env.example .env
```

3. Edit `.env` file with your database credentials:
```
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=seap
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h
OTP_EXPIRY=300
PORT=8081
```

4. Install Go dependencies:
```bash
go mod download
```

5. Run the backend (migrations run automatically):
```bash
go run main.go
```

The backend will start on `http://localhost:8081`

## Step 3: Frontend Setup

1. Open a new terminal and navigate to frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Create `.env.local` file:
```bash
echo "NEXT_PUBLIC_API_URL=http://localhost:8081" > .env.local
```

4. Run the frontend:
```bash
npm run dev
```

The frontend will start on `http://localhost:3000`

## Step 4: First Login

1. Open `http://localhost:3000` in your browser
2. Click "Register"
3. Create an account (you can choose "admin" role for admin access)
4. Check the backend console for the OTP code (in development mode)
5. Enter the OTP to verify your email
6. You'll be redirected to the dashboard

## Testing the Platform

### As a Regular User:
1. Create a campaign from "My Campaigns"
2. Fill in the campaign details
3. Submit for admin approval
4. Once approved, share the simulation link

### As an Admin:
1. Login with an admin account
2. Go to "Admin Panel" > "Campaigns"
3. Review and approve/reject campaigns
4. View analytics and leaderboard

### Testing a Simulation:
1. Get an approved campaign's simulation link
2. Visit: `http://localhost:3000/simulate/{token}`
3. Interact with the simulated phishing page
4. Submit the form (no real data is stored)
5. You'll be redirected to the awareness page

## Troubleshooting

### Backend won't start:
- Check PostgreSQL is running: `pg_isready`
- Verify database credentials in `.env`
- Check if port 8081 is available

### Frontend won't start:
- Make sure Node.js 18+ is installed: `node --version`
- Delete `node_modules` and run `npm install` again
- Check if port 3000 is available

### Database connection errors:
- Verify PostgreSQL is running
- Check database name, user, and password in `.env`
- Ensure database `seap` exists

### OTP not working:
- In development, OTP is shown in backend console and API response
- For production, configure email SMTP settings in `.env`

## Next Steps

- Configure email service for OTP delivery
- Set up proper JWT secret for production
- Configure CORS for your domain
- Set up SSL/TLS certificates
- Review and customize the awareness page content

## Development Tips

- Backend auto-reloads on file changes (if using `air` or similar)
- Frontend hot-reloads automatically
- Check browser console and backend logs for errors
- Use browser DevTools to inspect API calls

