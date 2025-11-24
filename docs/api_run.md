## Run API

#### Get Run

**Endpoint:** `GET /v1/runs/{id}`

**Path Parameters:**
- `id` (ULID) - Run identifier

**Response:**
```json
{
  "run": {
    "id": "01HQKXYZ123456789ABCDEFGHJK",
    "flow_id": "build-test",
    "namespace": "default",
    "status": "success",
    "trigger": "manual",
    "create_time": "2024-01-15T10:30:00Z",
    "start_time": "2024-01-15T10:30:05Z",
    "end_time": "2024-01-15T10:35:00Z",
    "variables": {
      "git_ref": "main"
    },
    "inline_run": {
      "job_id": "pipeline-runner-abc123",
      "steps": [
        {
          "id": "build",
          "status": "success",
          "exit_code": 0,
          "start_time": "2024-01-15T10:30:10Z",
          "end_time": "2024-01-15T10:32:00Z"
        },
        {
          "id": "test",
          "status": "success",
          "exit_code": 0,
          "start_time": "2024-01-15T10:32:05Z",
          "end_time": "2024-01-15T10:35:00Z"
        }
      ]
    }
  }
}
```

**Status Codes:**
- `200 OK` - Run found
- `404 Not Found` - Run doesn't exist

#### List Runs

**Endpoint:** `GET /v1/runs`

**Response:**
```json
{
  "runs": [
    {
      "id": "01HQKXYZ123456789ABCDEFGHJK",
      "namespace": "default",
      "flow_id": "build-test",
      "status": "success",
      "create_time": "2024-01-15T10:30:00Z"
    },
    {
      "id": "01HQKXY0987654321ZYXWVUTSRQP",
      "namespace": "default",
      "flow_id": "deploy-prod",
      "status": "running",
      "create_time": "2024-01-15T11:00:00Z"
    }
  ]
}
```

**Status Codes:**
- `200 OK` - List retrieved successfully

#### Cancel Run

**Endpoint:** `PUT /v1/runs/{id}/cancel`

**Path Parameters:**
- `id` (ULID) - Run identifier

**Response:**
```json
{}
```

**Status Codes:**
- `200 OK` - Run cancelled successfully
- `404 Not Found` - Run doesn't exist
- `409 Conflict` - Run cannot be cancelled (already completed)

#### Get Run Logs

**Endpoint:** `GET /v1/runs/{id}/logs`

**Path Parameters:**
- `id` (ULID) - Run identifier

**Query Parameters:**
- `job_id` (string) - Job ID (for inline runs)
- `step_id` (string) - Step ID to retrieve logs for
- `type` (string) - Log type: `stdout` or `stderr`
- `tail` (boolean) - Whether to stream logs (default: false)

**Response (tail=false):**
```json
{
  "logs": [
    "Starting build process...",
    "Compiling source files...",
    "Build completed successfully"
  ]
}
```

**Response (tail=true):**
Stream of log lines (newline-delimited)

**Status Codes:**
- `200 OK` - Logs retrieved successfully
- `404 Not Found` - Run or step doesn't exist
