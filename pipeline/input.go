package pipeline

import (
	"fmt"
	"strings"

	"github.com/TIBCOSoftware/flogo-lib/core/data"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/expression"
	"github.com/TIBCOSoftware/flogo-lib/core/mapper/exprmapper/expression/expr"
)


func NewInputValues(inputMetadata map[string]*data.Attribute, resolver data.Resolver, values map[string]interface{}) (*InputValues, error) {

	inputs := make([]InputValue, len(values))
	i := 0
	for name, value := range values {
		mdAttr, ok := inputMetadata[name]
		if !ok {
			return nil, fmt.Errorf("unknown input '%s'", name)
		}
		iv, err := NewInput(mdAttr, resolver, value)
		if err != nil {
			return nil, err
		}

		inputs[i] = iv
		i++
	}

	return &InputValues{inputs: inputs}, nil
}

type InputValues struct {
	inputs []InputValue
}

func (ivs *InputValues) GetAttrs(scope data.Scope) (map[string]*data.Attribute, error) {
	attrs := make(map[string]*data.Attribute, len(ivs.inputs))

	for _, iv := range ivs.inputs {
		attr, err := iv.GetValue(scope)
		if err != nil {
			return nil, err
		}
		attrs[attr.Name()] = attr
	}

	return attrs, nil
}

type InputValue interface {
	GetValue(scope data.Scope) (*data.Attribute, error)
}

func NewInput(mdAttr *data.Attribute, resolver data.Resolver, value interface{}) (InputValue, error) {

	if inVal, ok := value.(string); ok && strings.HasPrefix(inVal, "=") {
		//lets assume things that any resolver, assign or expression starts with "="

		strVal := inVal[1:]

		if isFixedResolver(strVal) {
			val, err := data.GetBasicResolver().Resolve(strVal, nil)
			if err != nil {
				return nil, err
			}
			return newFixedInputValue(mdAttr, val)
		}

		if isExpression(strVal) {
			return newExpressionInputValue(mdAttr, strVal, resolver)
		}

		if isAssign(strVal) {
			return newAssignInputValue(mdAttr, strVal, resolver)
		}

		if mdAttr.Type() == data.TypeString {
			//a string is expected, so probably a literal?
			return newFixedInputValue(mdAttr, inVal)
		}

		return nil, fmt.Errorf("invalid input value '%s' for type %s", inVal, data.TypeString.String())
	} else {
		//not a string, so isn't a mapping
		return newFixedInputValue(mdAttr, value)
	}
}

func isFixedResolver(val string) bool {
	return strings.HasPrefix(val, "$property") || strings.HasPrefix(val, "$env")
}

func isAssign(val string) bool {
	return strings.HasPrefix(val, "$")
}

func isExpression(val string) bool {
	return expression.IsExpression(val)
}

func newFixedInputValue(mdAttr *data.Attribute, value interface{}) (InputValue, error) {

	attr, err := data.NewAttribute(mdAttr.Name(), mdAttr.Type(), value)
	if err != nil {
		return nil, err
	}
	return &fixedInputValue{attr: attr}, nil
}

type fixedInputValue struct {
	attr *data.Attribute
}

func (iv *fixedInputValue) GetValue(scope data.Scope) (*data.Attribute, error) {
	return iv.attr, nil
}

func newAssignInputValue(mdAttr *data.Attribute, toResolve string, resolver data.Resolver) (InputValue, error) {

	return &assignInputValue{mdAttr: mdAttr, toResolve: toResolve, resolver: resolver}, nil
}

type assignInputValue struct {
	mdAttr    *data.Attribute
	toResolve string
	resolver  data.Resolver
}

func (iv *assignInputValue) GetValue(scope data.Scope) (*data.Attribute, error) {

	val, err := resolver.Resolve(iv.toResolve, scope)
	if err != nil {
		return nil, err
	}

	return data.NewAttribute(iv.mdAttr.Name(), iv.mdAttr.Type(), val)
}

func newExpressionInputValue(mdAttr *data.Attribute, exprStr string, resolver data.Resolver) (InputValue, error) {

	parsedExpr, err := expression.ParseExpression(exprStr)
	if err != nil {
		return nil, err
	}
	return &expressionInputValue{mdAttr: mdAttr, parsedExpr: parsedExpr, resolver: resolver}, nil
}

type expressionInputValue struct {
	mdAttr     *data.Attribute
	parsedExpr expr.Expr
	resolver   data.Resolver
}

func (iv *expressionInputValue) GetValue(scope data.Scope) (*data.Attribute, error) {

	val, err := iv.parsedExpr.EvalWithScope(scope, iv.resolver)
	if err != nil {
		return nil, err
	}

	return data.NewAttribute(iv.mdAttr.Name(), iv.mdAttr.Type(), val)
}
