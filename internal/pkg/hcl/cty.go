package hcl

import (
	"fmt"

	"github.com/oklog/ulid/v2"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/gocty"
)

func goValueToCtyValue(v any) (cty.Value, error) {
	if v == nil {
		return cty.NullVal(cty.DynamicPseudoType), nil
	}

	ty, err := gocty.ImpliedType(v)
	if err != nil {
		return GoToCty(v)
	}

	val, err := gocty.ToCtyValue(v, ty)
	if err != nil {
		return cty.NilVal, err
	}

	return val, nil
}

func GoToCty(v any) (cty.Value, error) {
	if v == nil {
		return cty.NullVal(cty.DynamicPseudoType), nil
	}

	switch val := v.(type) {
	case string:
		return cty.StringVal(val), nil
	case int:
		return cty.NumberIntVal(int64(val)), nil
	case int64:
		return cty.NumberIntVal(val), nil
	case float64:
		return cty.NumberFloatVal(val), nil
	case bool:
		return cty.BoolVal(val), nil
	case []any:
		vals := make([]cty.Value, len(val))
		for i, item := range val {
			itemVal, err := goValueToCtyValue(item)
			if err != nil {
				return cty.NilVal, err
			}
			vals[i] = itemVal
		}
		if len(vals) == 0 {
			return cty.ListValEmpty(cty.DynamicPseudoType), nil
		}
		return cty.TupleVal(vals), nil
	case map[string]any:
		vals := make(map[string]cty.Value)
		for k, item := range val {
			itemVal, err := goValueToCtyValue(item)
			if err != nil {
				return cty.NilVal, err
			}
			vals[k] = itemVal
		}
		if len(vals) == 0 {
			return cty.EmptyObjectVal, nil
		}
		return cty.ObjectVal(vals), nil
	case ulid.ULID:
		return cty.StringVal(val.String()), nil
	default:
		return cty.NilVal, fmt.Errorf("unsupported type: %T", v)
	}
}
