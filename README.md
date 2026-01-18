# Currency Conversion Service (GCP)

This project is a simple currency conversion service built to run on **Google Cloud Platform (GCP)**.

The initial goal is to provide a lightweight API that:
- Converts an amount from one currency to another
- Allows users (or systems) to save **favorite conversions** with predefined limits

This service is designed to be extended later with background jobs and additional flows.

## Features (Initial Scope)

### 1. Currency Conversion Endpoint
An API endpoint that receives:
- Origin currency (e.g. `USD`)
- Destination currency (e.g. `COP`)
- Amount

And returns the converted value using an exchange rate provider.

### 2. Favorite Conversion Endpoint
An API endpoint to store a **favorite conversion**, for example:
- Origin: `USD`
- Destination: `COP`
- Limit: a threshold amount or value of interest

These saved conversions are **not used immediately**, but are required for a future background process.

### 3. Background Job (Planned)
A scheduled **job** (not part of the initial API flow) will later:
- Read saved favorite conversions
- Perform additional logic (alerts, analysis, automation, etc.)

## Architecture (Early Stage)

- Single application
- Exposed via HTTP API
- Designed to run on GCP (compute, storage, and jobs to be defined)
- Persistence layer for favorite conversions

## Status

🚧 **Early development**  
This repository currently focuses on defining the API boundaries and core behavior. More details will be added as the implementation evolves.

