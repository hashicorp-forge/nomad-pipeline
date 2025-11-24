package coordinator

import (
	"fmt"
	"strings"

	"github.com/hashicorp-forge/nomad-pipeline/internal/pkg/state"
)

// generateVariablesMap is a helper function that generates a map of variables
// for the flow run. It starts with the default values defined in the flow's
// variables and then overrides them with any provided runtime vars. If any
// required variables are missing, an error is returned.
func generateVariablesMap(flow *state.Flow, vars map[string]any) (map[string]any, error) {

	var errors []error

	varNS, ok := vars["var"].(map[string]any)
	if !ok {
		varNS = vars
	}

	// Initialize the result map; we do not know the size yet.
	result := make(map[string]any)

	for _, v := range flow.Variables {

		nameSplit := strings.Split(v.Name, ".")

		switch len(nameSplit) {
		case 2:
			namespace := nameSplit[0]
			name := nameSplit[1]

			// Ensure the namespace map exists in the result.
			if _, ok := result[namespace]; !ok {
				result[namespace] = make(map[string]any)
			}
			nsMap := result[namespace].(map[string]any)

			// Set any default values from flow variables first, so that if they
			// are not present in the runtime vars, they will still be set.
			if v.Default != nil {
				nsMap[name] = v.Default
			}

			// If a define flow variable is provided in the runtime vars, override
			// the default value or set it if not present.
			if runtimeVar, ok := varNS[namespace]; ok {
				runtimeNsMap, ok := runtimeVar.(map[string]any)
				if !ok {
					errors = append(errors, fmt.Errorf("invalid variable namespace: %s", namespace))
					continue
				}
				if val, ok := runtimeNsMap[name]; ok {
					nsMap[name] = val
				}
			}

			// Check for required variables that are missing.
			if runtimeVar, ok := varNS[namespace]; ok {
				runtimeNsMap, ok := runtimeVar.(map[string]any)
				if !ok {
					errors = append(errors, fmt.Errorf("invalid variable namespace: %s", namespace))
					continue
				}
				if _, exists := runtimeNsMap[name]; !exists && v.Required {
					errors = append(errors, fmt.Errorf("missing required variable: %s.%s", namespace, name))
				}
			} else if v.Required {
				errors = append(errors, fmt.Errorf("missing required variable: %s.%s", namespace, name))
			}

			result[namespace] = nsMap

		case 1:
			name := nameSplit[0]

			// Set any default values from flow variables first, so that if they
			// are not present in the runtime vars, they will still be set.
			if v.Default != nil {
				result[name] = v.Default
			}

			// If a define flow variable is provided in the runtime vars, override
			// the default value or set it if not present.
			if runtimeVar, ok := varNS[name]; ok {
				result[name] = runtimeVar
			}

			// Check for required variables that are missing.
			if _, exists := varNS[name]; !exists && v.Required {
				errors = append(errors, fmt.Errorf("missing required variable: %s", name))
			}
		default:
			errors = append(errors, fmt.Errorf("invalid variable name: %s", v.Name))
			continue
		}
	}

	if len(errors) > 0 {
		return nil, fmt.Errorf("variable errors: %v", errors)
	}

	// Return the final variables map under the "var" key as this is standard
	// throughout HCL use.
	return map[string]any{"var": result}, nil
}
