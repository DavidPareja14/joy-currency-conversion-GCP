# Currency Conversion Service

A microservices-based currency conversion system built on Google Cloud Platform that monitors exchange rates and sends email notifications when user-defined thresholds are exceeded.

## Architecture Overview

```
Cloud Scheduler
         ↓
    Worker (Cloud Run)
         ↓
   [Cloud SQL] + [Exchange Rates API]
         ↓
    Pub/Sub Topic
         ↓
  Cloud Function → Email
```

## Project Structure

```
.
├── api/              # Main API service (Compute Engine)
├── worker/           # Threshold checker (Cloud Run)
├── function/         # Email notification service (Cloud Functions)
└── docker-compose.yml
```

---

## 1. API Service

**Location**: Root directory (`main.go`, `logic.go`, `config/`)  
**Deployment**: Compute Engine (Docker Compose)  
**Database**: Cloud SQL (MySQL with private IP)

### Endpoints

#### `GET /convert`
Fetches current EUR to COP exchange rate.

**Response**:
```json
{
  "from": "EUR",
  "to": "COP",
  "rate": 4523.45,
  "date": "2026-01-26",
  "timestamp": 1706227200
}
```

#### `POST /favorites`
Saves a favorite currency conversion with a threshold.

**Request Body**:
```json
{
  "email": "user@example.com",
  "currency_origin": "EUR",
  "currency_destination": "COP",
  "threshold": 4500.0
}
```

**Response**:
```json
{
  "id": 1,
  "email": "user@example.com",
  "currency_origin": "EUR",
  "currency_destination": "COP",
  "threshold": 4500.0
}
```

**Notes**:
- Only one favorite per email (unique constraint)
- Currency origin must be "EUR"
- Returns 409 Conflict if email already exists

---

## 2. Worker Service

**Location**: `worker/`  
**Deployment**: Cloud Run  
**Trigger**: Cloud Scheduler

### Endpoints

#### `POST /check-thresholds`
Checks all favorite conversions and publishes notifications to Pub/Sub when thresholds are exceeded.

**Response**:
```json
{
  "message": "thresholds checked successfully",
  "status": "success"
}
```

**Process**:
1. Fetches all favorites from Cloud SQL
2. Calls Exchange Rates API for each
3. Compares current rate with threshold
4. Publishes notification to Pub/Sub if threshold exceeded

#### `DELETE /delete-all-favorites`
Deletes all favorite conversions from the database.

**Response**:
```json
{
  "message": "all favorites deleted",
  "rows_affected": 5
}
```

#### `GET /health`
Health check endpoint.

**Response**: `200 OK`

---

## 3. Email Function

**Location**: `function/`  
**Deployment**: Cloud Functions  
**Trigger**: Pub/Sub subscription (Push)  
**Runtime**: Python 3.12

### Trigger Format

Receives Pub/Sub messages with the following data:

```json
{
  "email": "user@example.com",
  "currency_origin": "EUR",
  "currency_destination": "COP",
  "current_rate": 4550.0,
  "threshold": 4500.0
}
```

**Action**: Sends email notification via Gmail SMTP.

---

## GCP Services Used

- **Cloud SQL**: MySQL database with private IP
- **Cloud Run**: Serverless worker container
- **Cloud Functions**: Serverless email sender
- **Pub/Sub**: Asynchronous messaging
- **Secret Manager**: Secure credential storage
- **Cloud Scheduler**: Automated job execution
- **VPC Connector**: Private network connectivity
- **Artifact Registry**: Container image storage

---

## Environment Variables

### API Service
- `PORT`: Server port (default: 8080)
- `GCP_PROJECT_ID`: GCP project ID
- `DB_NAME`: Database name
- `CLOUD_SQL_CONNECTION_NAME`: Cloud SQL instance connection name

### Worker Service
- `PORT`: Server port (default: 8081)
- `GCP_PROJECT_ID`: GCP project ID
- `DB_NAME`: Database name
- `CLOUD_SQL_CONNECTION_NAME`: Cloud SQL instance connection name

### Secrets (Secret Manager)
- `EXCHANGE_RATES_API_KEY`: API key for exchangeratesapi.io
- `DB_USER`: Database username
- `DB_PASSWORD`: Database password
- `PUBSUB_TOPIC_ID`: Pub/Sub topic name
- `SMTP_USER`: Gmail address
- `SMTP_PASSWORD`: Gmail app password

---

## Quick Start (Local)

```bash
# Start all services
docker-compose up

# API available at: http://localhost:8080
# Worker available at: http://localhost:8081
```

---

## Documentation

For detailed setup and deployment instructions, see the [Wiki](../../wiki):

- [API Service Setup](../../wiki/API-Service)
- [Worker Setup](../../wiki/Worker-Service)
- [Function Setup](../../wiki/Email-Function)
- [GCP Configuration](../../wiki/GCP-Configuration)
- [Local Development](../../wiki/Local-Development)

