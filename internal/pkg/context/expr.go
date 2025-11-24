package context

import (
	"fmt"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"

	hhcl "github.com/hashicorp-forge/nomad-pipeline/internal/pkg/hcl"
)

func (c *Context) ParseBoolExpr(expr string) (bool, error) {
	evalCtx, err := c.createEvalContext()
	if err != nil {
		return false, fmt.Errorf("failed to create eval context: %w", err)
	}

	parsedExpr, diags := hclsyntax.ParseExpression([]byte(expr), "<expr>", hcl.InitialPos)
	if diags.HasErrors() {
		return false, fmt.Errorf("failed to parse condition: %s", diags.Error())
	}

	val, diags := parsedExpr.Value(evalCtx)
	if diags.HasErrors() {
		return false, fmt.Errorf("failed to evaluate condition: %s", diags.Error())
	}

	if val.Type() != cty.Bool {
		return false, fmt.Errorf("condition must evaluate to boolean, got %s",
			val.Type().FriendlyName())
	}

	return val.True(), nil
}

func (c *Context) ParseTemplateStringExpr(expr string) (string, error) {

	evalCtx, err := c.createEvalContext()
	if err != nil {
		return "", fmt.Errorf("failed to create eval context: %w", err)
	}

	parsedExpr, diags := hclsyntax.ParseTemplate([]byte(expr), "<tpl>", hcl.InitialPos)
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to parse expression: %s", diags.Error())
	}

	val, diags := parsedExpr.Value(evalCtx)
	if diags.HasErrors() {
		return "", fmt.Errorf("failed to evaluate expression: %s", diags.Error())
	}

	if val.Type() != cty.String {
		return "", fmt.Errorf("expression must evaluate to string, got %s",
			val.Type().FriendlyName())
	}

	return val.AsString(), nil

}

func (c *Context) createEvalContext() (*hcl.EvalContext, error) {

	contextMap := c.AsMap()

	variables := make(map[string]cty.Value)

	for key, value := range contextMap {
		ctyVal, err := hhcl.GoToCty(value)
		if err != nil {
			return nil, fmt.Errorf("failed to convert %s to cty value: %w", key, err)
		}
		variables[key] = ctyVal
	}

	return &hcl.EvalContext{
		Variables: variables,
		Functions: hclFuncs(),
	}, nil
}

func hclFuncs() map[string]function.Function {
	return map[string]function.Function{
		"always": alwaysHCLFunc(),
	}
}

// always is the HCL function that always returns true and can be called via
// always().
func alwaysHCLFunc() function.Function {
	return function.New(&function.Spec{
		Params: nil,
		Type:   func(args []cty.Value) (cty.Type, error) { return cty.Bool, nil },
		Impl:   func(args []cty.Value, retType cty.Type) (cty.Value, error) { return cty.BoolVal(true), nil },
	})
}
