## Trigger API

#### Create Trigger

**Endpoint:** `POST /v1/triggers`

**Request:**
```json
{
  "trigger": {
    "id": "terraform-provider-nomad",
    "namespace": "jrasell",
    "flow": "terraform-provider-nomad",
    "source": {
      "id": "github-push",
      "provider": "git-webhook",
      "config": "provider = \"github\"\nrepository = \"jrasell/terraform-provider-nomad\"\nevents = [\"push\"]"
    }
  }
}
```

**Response:**
```json
{
  "trigger": {
    "id": "terraform-provider-nomad",
    "namespace": "jrasell",
    "flow": "terraform-provider-nomad",
    "source": {
      "id": "github-push",
      "provider": "git-webhook",
      "config": "provider = \"github\"\nrepository = \"jrasell/terraform-provider-nomad\"\nevents = [\"push\"]"
    }
  }
}
```

**Status Codes:**
- `200 OK` - Trigger created successfully
- `400 Bad Request` - Invalid trigger definition
- `409 Conflict` - Trigger already exists

#### Get Trigger

**Endpoint:** `GET /v1/triggers/{id}`

**Path Parameters:**
- `id` (string) - Trigger identifier

**Response:**
```json
{
  "trigger": {
    "id": "terraform-provider-nomad",
    "namespace": "jrasell",
    "flow": "terraform-provider-nomad",
    "source": {
      "id": "github-push",
      "provider": "git-webhook",
      "config": "provider = \"github\"\nrepository = \"jrasell/terraform-provider-nomad\"\nevents = [\"push\"]"
    }
  }
}
```

**Status Codes:**
- `200 OK` - Trigger found
- `404 Not Found` - Trigger doesn't exist

#### List Triggers

**Endpoint:** `GET /v1/triggers`

**Response:**
```json
{
  "triggers": [
    {
      "id": "terraform-provider-nomad",
      "namespace": "jrasell",
      "flow": "terraform-provider-nomad"
    },
    {
      "id": "job-cascade-scheduled",
      "namespace": "default",
      "flow": "job-cascade"
    }
  ]
}
```

**Status Codes:**
- `200 OK` - List retrieved successfully

#### Delete Trigger

**Endpoint:** `DELETE /v1/triggers/{id}`

**Path Parameters:**
- `id` (string) - Trigger identifier

**Response:**
```json
{}
```

**Status Codes:**
- `200 OK` - Trigger deleted successfully
- `404 Not Found` - Trigger doesn't exist
