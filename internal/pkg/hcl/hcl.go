package hcl

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

type EvalContext = hcl.EvalContext

func GenerateEvalContext(vars map[string]any) (*hcl.EvalContext, error) {
	ctx := &hcl.EvalContext{
		Variables: make(map[string]cty.Value),
		Functions: hclFunctions(),
	}

	if len(vars) == 0 {
		return ctx, nil
	}

	varValues := make(map[string]cty.Value)

	for k, v := range vars {
		ctyVal, err := goValueToCtyValue(v)
		if err != nil {
			return nil, fmt.Errorf("failed to convert variable %q: %w", k, err)
		}
		varValues[k] = ctyVal
	}

	ctx.Variables = varValues

	return ctx, nil
}

func EvaluateTemplateString(tpl string, evalCtx *hcl.EvalContext) (string, error) {

	expr, diags := hclsyntax.ParseTemplate([]byte(tpl), "<template>", hcl.InitialPos)
	if diags.HasErrors() {
		return "", fmt.Errorf("parse error: %s", diags.Error())
	}

	val, diags := expr.Value(evalCtx)
	if diags.HasErrors() {
		return "", fmt.Errorf("eval error: %s", diags.Error())
	}

	if !val.Type().Equals(cty.String) {
		return "", fmt.Errorf("result is not a string: %s", val.Type().FriendlyName())
	}

	return val.AsString(), nil
}
