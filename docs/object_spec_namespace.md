# Namespace Specification

A Namespace provides logical isolation and organization for Flows, Runs, and Triggers within Nomad
Pipeline. Namespaces allow teams to manage their pipelines independently while sharing the same
Nomad Pipeline server instance.

### Namespace Attributes

- `id` (string, required): A unique identifier for the namespace. This is used to reference the
  namespace in API calls and configurations.
  
- `description` (string, optional): A human-readable description of the namespace.

### Examples

A simple namespace definition in HCL format:
```hcl
namespace {
  id          = "namespace-name"
  description = "Optional description"
}
```

A simple namespace definition in JSON format:
```json
{
  "namespace": {
    "id": "namespace-name",
    "description": "Optional description"
  }
}
```
