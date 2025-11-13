# SEAP - Social Engineering Education Platform

A comprehensive platform for educating users about phishing and social engineering through safe, interactive simulations.

## Features

- **User Authentication**: JWT-based authentication with OTP email verification
- **Campaign Management**: Create, manage, and track phishing simulation campaigns
- **Admin Panel**: Approve/reject campaigns, manage users, view analytics
- **Phishing Simulation**: Safe simulation system with unique tracking links
- **Awareness Pages**: Educational content after simulated phishing attempts
- **Analytics Dashboard**: Comprehensive analytics with Chart.js visualizations
- **Leaderboard**: Gamification system ranking top campaign creators
- **Event Tracking**: Detailed logging of all user interactions

## Technology Stack

### Backend
- **Go** with `gorilla/mux` router
- **PostgreSQL** database
- **JWT** for authentication
- **OTP** for email verification

### Frontend
- **Next.js 14** with TypeScript
- **Tailwind CSS** for styling
- **Chart.js** for data visualization
- **React Hot Toast** for notifications

## Project Structure

```
.
├── backend/
│   ├── main.go
│   ├── go.mod
│   └── internal/
│       ├── auth/          # Authentication utilities
│       ├── database/       # Database connection and migrations
│       ├── handlers/      # HTTP handlers
│       ├── middleware/    # Auth and CORS middleware
│       ├── models/        # Data models
│       └── utils/         # Utility functions
├── frontend/
│   ├── app/              # Next.js app directory
│   ├── components/       # React components
│   └── lib/             # Utilities and API client
└── README.md
```

## Setup Instructions

### Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- PostgreSQL 12 or higher

### Backend Setup

1. Navigate to the backend directory:
```bash
cd backend
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file (copy from `.env.example`):
```bash
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=seap
JWT_SECRET=your-secret-key-change-in-production
JWT_EXPIRY=24h
OTP_EXPIRY=300
PORT=8081
```

4. Create the PostgreSQL database:
```bash
createdb seap
```

5. Run the backend:
```bash
go run main.go
```

The backend will automatically run migrations on startup.

### Frontend Setup

1. Navigate to the frontend directory:
```bash
cd frontend
```

2. Install dependencies:
```bash
npm install
```

3. Create a `.env.local` file:
```bash
NEXT_PUBLIC_API_URL=http://localhost:8081
```

4. Run the development server:
```bash
npm run dev
```

The frontend will be available at `http://localhost:3000`

## Usage

### Creating a Campaign

1. Register/Login as a user
2. Navigate to "My Campaigns"
3. Click "Create Campaign"
4. Fill in the campaign details:
   - Title
   - Description
   - Email text (simulated email content)
   - Landing page URL (optional)
   - Expiry date (optional)
5. Submit for admin approval

### Admin Workflow

1. Login as an admin user
2. Navigate to "Admin Panel" > "Campaigns"
3. Review pending campaigns
4. Approve or reject with comments
5. View analytics and leaderboard

### Simulating a Phishing Attempt

1. Once a campaign is approved, get the simulation link
2. Share the link: `http://localhost:3000/simulate/{token}`
3. When someone interacts with the simulation:
   - Link opened event is logged
   - Form submission is logged
   - User is redirected to awareness page
   - Awareness page view is logged

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register new user
- `POST /api/auth/login` - Login user
- `POST /api/auth/verify-otp` - Verify OTP code
- `POST /api/auth/resend-otp` - Resend OTP

### User
- `GET /api/user/profile` - Get user profile
- `PUT /api/user/profile` - Update profile
- `GET /api/user/campaigns` - Get user campaigns
- `POST /api/user/campaigns` - Create campaign
- `GET /api/user/campaigns/{id}` - Get campaign
- `PUT /api/user/campaigns/{id}` - Update campaign
- `DELETE /api/user/campaigns/{id}` - Delete campaign
- `GET /api/user/analytics` - Get user analytics

### Admin
- `GET /api/admin/campaigns` - Get all campaigns
- `POST /api/admin/campaigns/{id}/approve` - Approve campaign
- `POST /api/admin/campaigns/{id}/reject` - Reject campaign
- `GET /api/admin/users` - Get all users
- `DELETE /api/admin/users/{id}` - Delete user
- `GET /api/admin/audit-logs` - Get audit logs
- `GET /api/admin/leaderboard` - Get leaderboard
- `GET /api/admin/analytics` - Get admin analytics

### Simulation
- `GET /api/simulate/{token}` - Get simulation landing page
- `POST /api/simulate/{token}/submit` - Submit simulation form
- `GET /api/awareness/{token}` - Get awareness page

## Security Features

- JWT-based authentication
- Password hashing with bcrypt
- OTP email verification
- Admin approval required for campaigns
- No sensitive data storage
- Input sanitization
- CORS protection
- Audit logging for admin actions

## Development

### Running Tests

Backend tests (when implemented):
```bash
cd backend
go test ./...
```

Frontend tests (when implemented):
```bash
cd frontend
npm test
```

### Building for Production

Backend:
```bash
cd backend
go build -o seap-server main.go
```

Frontend:
```bash
cd frontend
npm run build
npm start
```

## License

This project is for educational purposes only.

## Contributing

This is an educational project. Feel free to fork and modify for your own learning purposes.

## Notes

- OTP codes are currently returned in API responses for development. Remove this in production.
- Email sending functionality is not implemented. Integrate with an email service (SendGrid, AWS SES, etc.) for production.
- The platform is designed to be safe and ethical - no real credentials are stored or used maliciously.

