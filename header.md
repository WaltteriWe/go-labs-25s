# Go Labs API Documentation

Welcome to the Go Labs API documentation. This API provides endpoints for managing articles and other resources.

## Base URL

All API requests should be made to: `http://localhost:8080`

## Content Type

All requests and responses use `application/json` content type unless otherwise specified.

## Getting Started

1. Start the Go server: `go run main.go`
2. The API will be available at `http://localhost:8080`
3. Use the endpoints documented below to interact with the API

## Authentication

Currently, this API does not require authentication. All endpoints are publicly accessible.

## Error Handling

The API uses standard HTTP status codes:

- `200` - Success
- `201` - Created  
- `400` - Bad Request
- `404` - Not Found
- `500` - Internal Server Error

Error responses will include a JSON object with error details.