# API

The main interface to Nomad Pipeline is a RESTful HTTP API. The API can query the current state of
the system as well as modify state object and processes.

## Version Prefix
All API endpoints are prefixed with a version identifier to allow for future enhancements and
backwards compatibility. The current version prefix is `v1`.

## Namespaces
Namespaces provide logical isolation and organization for Flows, Runs, and Triggers within Nomad
Pipeline. When using the non-default `default` namespace, the API request must pass the target
namespace as a `namespace` API query parameter.

Here is an example using curl to query the `staging` namespace:
```bash
curl -X GET "http://localhost:8080/v1/flows?namespace=staging"
```

List endpoints support the wildcard namespace identifier `*` to list resources across all
namespaces:
```bash
curl -X GET "http://localhost:8080/v1/flows?namespace=*"
```
