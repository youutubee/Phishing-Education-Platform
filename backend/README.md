# SEAP Backend

Backend API for the SEAP (Social Engineering Awareness Platform) project.

## Prerequisites

- Go 1.21 or higher
- MongoDB (local or MongoDB Atlas)

## Setup Instructions

### 1. Install Dependencies

```bash
cd backend
go mod download
```

### 2. Configure Environment Variables

Create a `.env` file in the `backend` directory:

```bash
cp .env.example .env
```

Edit the `.env` file with your configuration:

```env
# MongoDB Configuration
MONGODB_URL=mongodb://localhost:27017
# OR for MongoDB Atlas:
# MONGODB_URL=mongodb+srv://username:password@cluster.mongodb.net/?retryWrites=true&w=majority

# Database Name
DB_NAME=seap

# Server Port
PORT=8081

# JWT Secret (generate a strong random key for production)
JWT_SECRET=your-super-secret-jwt-key-change-this-in-production

# JWT Expiry
JWT_EXPIRY=24h
```

### 3. MongoDB Setup

#### Option A: Local MongoDB

1. Install MongoDB locally
2. Start MongoDB service
3. Use connection string: `mongodb://localhost:27017`

#### Option B: MongoDB Atlas (Cloud)

1. Create a free account at [MongoDB Atlas](https://www.mongodb.com/cloud/atlas)
2. Create a new cluster
3. Create a database user
4. Whitelist your IP address in Network Access
5. Get your connection string from "Connect" → "Connect your application"
6. Use the connection string in `MONGODB_URL`

**Important for Atlas:**
- Connection string should start with `mongodb+srv://`
- Replace `<password>` with your actual password
- Ensure your IP is whitelisted in Network Access

### 4. Run the Backend

```bash
# From the backend directory
go run main.go
```

Or build and run:

```bash
# Build
go build -o seap-backend main.go

# Run
./seap-backend
```

The server will start on `http://localhost:8081` (or the port specified in `.env`)

## Environment Variables

| Variable | Description | Required | Default |
|----------|-------------|----------|---------|
| `MONGODB_URL` | MongoDB connection string | Yes | `mongodb://localhost:27017` |
| `DB_NAME` | Database name | No | `seap` |
| `PORT` | Server port | No | `8081` |
| `JWT_SECRET` | Secret key for JWT tokens | Yes | - |
| `JWT_EXPIRY` | JWT token expiry duration | No | `24h` |
| `RESEND_API_KEY` | API key for Resend transactional emails | No | - |
| `RESEND_FROM_EMAIL` | Sender address used for Resend emails | No | `no-reply@seap.local` |
| `APP_BASE_URL` | Base URL of the frontend (used in emails) | No | `http://localhost:3000` |

## API Endpoints

### Health Check
- `GET /api/health` - Check if server is running

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login user
- `POST /api/auth/verify-otp` - Verify OTP code
- `POST /api/auth/resend-otp` - Resend OTP code

### User Routes (Protected)
- `GET /api/user/profile` - Get user profile
- `PUT /api/user/profile` - Update user profile
- `POST /api/user/campaigns` - Create a campaign
- `GET /api/user/campaigns` - Get user's campaigns
- `GET /api/user/campaigns/{id}` - Get specific campaign
- `PUT /api/user/campaigns/{id}` - Update campaign
- `DELETE /api/user/campaigns/{id}` - Delete campaign
- `GET /api/user/analytics` - Get user analytics
- `GET /api/leaderboard` - Get leaderboard (available to all authenticated users)

### Admin Routes (Protected - Admin Only)
- `GET /api/admin/campaigns` - Get all campaigns
- `POST /api/admin/campaigns/{id}/approve` - Approve campaign
- `POST /api/admin/campaigns/{id}/reject` - Reject campaign
- `GET /api/admin/users` - Get all users
- `DELETE /api/admin/users/{id}` - Delete user
- `GET /api/admin/audit-logs` - Get audit logs
- `GET /api/admin/analytics` - Get admin analytics
- `GET /api/admin/leaderboard` - Get leaderboard (admin-specific entry point)

### Simulation Routes (Public)
- `GET /api/simulate/{token}` - Simulate landing page
- `POST /api/simulate/{token}/submit` - Submit simulated form
- `GET /api/awareness/{token}` - Show awareness page

## Troubleshooting

### Connection Issues

1. **MongoDB Connection Failed**
   - Check if MongoDB is running (local) or accessible (Atlas)
   - Verify connection string format
   - For Atlas: Ensure IP is whitelisted
   - Check firewall settings

2. **Port Already in Use**
   - Change `PORT` in `.env` to a different port
   - Or stop the process using port 8081

3. **TLS Errors (Atlas)**
   - Ensure connection string uses `mongodb+srv://`
   - Check network connectivity
   - Verify credentials are correct

## Development

### Run with Hot Reload (using air)

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with air
air
```

### Build for Production

```bash
go build -o seap-backend -ldflags="-s -w" main.go
```

## License

MIT

