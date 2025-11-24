## Namespace API

#### Create Namespace

**Endpoint:** `POST /v1/namespaces`

**Request:**
```json
{
  "namespace": {
    "id": "team-alpha",
    "description": "Alpha team's pipeline namespace"
  }
}
```

**Response:**
```json
{
  "namespace": {
    "id": "team-alpha",
    "description": "Alpha team's pipeline namespace"
  }
}
```

**Status Codes:**
- `200 OK` - Namespace created successfully
- `400 Bad Request` - Invalid namespace definition
- `409 Conflict` - Namespace already exists

#### Get Namespace

**Endpoint:** `GET /v1/namespaces/{name}`

**Path Parameters:**
- `name` (string) - Namespace identifier

**Response:**
```json
{
  "namespace": {
    "id": "team-alpha",
    "description": "Alpha team's pipeline namespace"
  }
}
```

**Status Codes:**
- `200 OK` - Namespace found
- `404 Not Found` - Namespace doesn't exist

#### List Namespaces

**Endpoint:** `GET /v1/namespaces`

**Response:**
```json
{
  "namespaces": [
    {
      "id": "default",
      "description": "Default namespace"
    },
    {
      "id": "team-alpha",
      "description": "Alpha team's pipeline namespace"
    },
    {
      "id": "production",
      "description": "Production pipelines"
    }
  ]
}
```

**Status Codes:**
- `200 OK` - List retrieved successfully

#### Delete Namespace

**Endpoint:** `DELETE /v1/namespaces/{name}`

**Path Parameters:**
- `name` (string) - Namespace identifier

**Response:**
```json
{}
```

**Status Codes:**
- `200 OK` - Namespace deleted successfully
- `404 Not Found` - Namespace doesn't exist
- `409 Conflict` - Namespace contains resources (flows, runs)
