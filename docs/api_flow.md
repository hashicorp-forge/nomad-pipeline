## Flow API

#### Create Flow

**Endpoint:** `POST /v1/flows`

**Request:**
```json
{
  "flow": {
    "id": "build-test",
    "namespace": "default",
    "variables": [
      {
        "name": "git_ref",
        "type": "string",
        "required": true
      }
    ],
    "inline": {
      "id": "ci-pipeline",
      "runner": {
        "nomad_on_demand": {
          "namespace": "default",
          "image": "golang:1.21",
          "artifacts": [
            {
              "source": "git::https://github.com/org/repo",
              "destination": "src",
              "options": {
                "ref": "${var.git_ref}"
              }
            }
          ],
          "resource": {
            "cpu": 2000,
            "memory": 2048
          }
        }
      },
      "steps": [
        {
          "id": "build",
          "run": "cd src\ngo build -v ./..."
        }
      ]
    }
  }
}
```

**Response:**
```json
{
  "flow": {
    "id": "build-test",
    "namespace": "default",
    "variables": [
      {
        "name": "git_ref",
        "type": "string",
        "required": true
      }
    ],
    "inline": {
      "id": "ci-pipeline",
      "runner": {
        "nomad_on_demand": {
          "namespace": "default",
          "image": "golang:1.21",
          "artifacts": [
            {
              "source": "git::https://github.com/org/repo",
              "destination": "src",
              "options": {
                "ref": "${var.git_ref}"
              }
            }
          ],
          "resource": {
            "cpu": 2000,
            "memory": 2048
          }
        }
      },
      "steps": [
        {
          "id": "build",
          "run": "cd src\ngo build -v ./..."
        }
      ]
    }
  }
}
```

**Status Codes:**
- `200 OK` - Flow created successfully
- `400 Bad Request` - Invalid flow definition
- `409 Conflict` - Flow already exists

#### Get Flow

**Endpoint:** `GET /v1/flows/{id}`

**Path Parameters:**
- `id` (string) - Flow identifier

**Response:**
```json
{
  "flow": {
    "id": "build-test",
    "namespace": "default",
    "variables": [
      {
        "name": "git_ref",
        "type": "string",
        "required": true
      }
    ],
    "inline": {
      "id": "ci-pipeline",
      "runner": {
        "nomad_on_demand": {
          "namespace": "default",
          "image": "golang:1.21",
          "artifacts": [],
          "resource": {
            "cpu": 2000,
            "memory": 2048
          }
        }
      },
      "steps": [
        {
          "id": "build",
          "run": "cd src\ngo build -v ./..."
        }
      ]
    }
  }
}
```

**Status Codes:**
- `200 OK` - Flow found
- `404 Not Found` - Flow doesn't exist

#### List Flows

**Endpoint:** `GET /v1/flows`

**Response:**
```json
{
  "flows": [
    {
      "id": "build-test",
      "namespace": "default",
      "type": "inline"
    },
    {
      "id": "deploy-pipeline",
      "namespace": "production",
      "type": "specification"
    }
  ]
}
```

**Status Codes:**
- `200 OK` - List retrieved successfully

#### Delete Flow

**Endpoint:** `DELETE /v1/flows/{id}`

**Path Parameters:**
- `id` (string) - Flow identifier

**Response:**
```json
{}
```

**Status Codes:**
- `200 OK` - Flow deleted successfully
- `404 Not Found` - Flow doesn't exist

#### Run Flow

**Endpoint:** `POST /v1/flows/{id}/run`

**Path Parameters:**
- `id` (string) - Flow identifier

**Request:**
```json
{
  "id": "build-test",
  "variables": {
    "git_ref": "main"
  }
}
```

**Response:**
```json
{
  "run_id": "01HQKXYZ123456789ABCDEFGHJK"
}
```

**Status Codes:**
- `200 OK` - Flow run initiated successfully
- `400 Bad Request` - Invalid variables or flow configuration
- `404 Not Found` - Flow doesn't exist
