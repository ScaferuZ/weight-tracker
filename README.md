# Weight Tracker

A private, self-hosted weight tracking application built with Go, HTMX, and SQLite.

## Overview

This minimalist weight tracking application addresses the common pain points of commercial weight trackers:
- **Data Privacy**: Self-hosted with local SQLite storage
- **No Subscriptions**: Free and open source
- **Feature Simplicity**: Focused on core weight tracking functionality

## Technology Stack

- **Backend**: Go 1.22+ with net/http standard library
- **Frontend**: HTML with HTMX for dynamic interactions
- **Database**: SQLite for embedded, serverless storage
- **Charts**: Chart.js for weight progression visualization
- **Deployment**: Docker container with OpenMediaVault compatibility
- **Styling**: Tailwind CSS for responsive design

## Features

- User registration and authentication
- Daily weight logging with automatic updates
- Weight history with pagination
- Interactive weight progression chart
- Basic statistics (current weight, changes over time)
- Mobile-responsive design
- Data export functionality

## Project Structure

```
weight-tracker/
├── cmd/server/main.go              # Application entry point
├── internal/
│   ├── handlers/                   # HTTP request handlers
│   ├── models/                     # Data models and database logic
│   ├── middleware/                 # HTTP middleware
│   └── config/                     # Configuration management
├── static/                         # Static assets
├── templates/                      # HTML templates
├── migrations/                     # Database migrations
├── docker-compose.yml             # Development deployment
├── Dockerfile                     # Production container
└── README.md                      # Documentation
```

## Installation

### Docker (Recommended)

```bash
docker-compose up -d
```

### Local Development

```bash
go run ./cmd/server
```

Access the application at http://localhost:8080

## License

MIT License