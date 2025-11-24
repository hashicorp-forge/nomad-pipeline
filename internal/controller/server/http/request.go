package http

import "net/http"

const namespaceQueryParam = "namespace"

func getNamespaceParam(r *http.Request) string { return r.URL.Query().Get(namespaceQueryParam) }
