.PHONY: help setup-backend setup-frontend run-backend run-frontend run-all

help:
	@echo "Available commands:"
	@echo "  make setup-backend   - Install backend dependencies"
	@echo "  make setup-frontend - Install frontend dependencies"
	@echo "  make run-backend    - Run backend server"
	@echo "  make run-frontend   - Run frontend server"
	@echo "  make run-all         - Run both backend and frontend"

setup-backend:
	@echo "Setting up backend..."
	cd backend && go mod download

setup-frontend:
	@echo "Setting up frontend..."
	cd frontend && npm install

run-backend:
	@echo "Starting backend server..."
	cd backend && go run main.go

run-frontend:
	@echo "Starting frontend server..."
	cd frontend && npm run dev

run-all: run-backend run-frontend

