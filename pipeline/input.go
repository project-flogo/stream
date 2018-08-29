package pipeline

import (
	"fmt"
	"strings"

	"github.com/flogo-oss/core/core/mapper/exprmapper/expression"
	"github.com/flogo-oss/core/core/mapper/exprmapper/expression/expr"
	"github.com/flogo-oss/core/data"
)

type DetailedAttribute struct {
	*data.Attribute

	isNew bool
}

func NewInputValues(inputMetadata map[string]*data.Attribute, resolver data.Resolver, values map[string]interface{}, ignoreNew bool) (*InputValues, error) {

	inputs := make([]InputValue, len(values))
	i := 0
	newAttr := false

	for name, value := range values {
		mdAttr, ok := inputMetadata[name]
		var attrName string
		var attrType data.Type

		if !ok {
			if !ignoreNew {
				return nil, fmt.Errorf("unknown input '%s'", name)
			}
			attrName, attrType = getAttrInfo(name)
			newAttr = true
		} else {
			attrName = mdAttr.Name()
			attrType = mdAttr.Type()
		}
		iv, err := NewInput(attrName, attrType, resolver, value, newAttr)
		if err != nil {
			return nil, err
		}

		inputs[i] = iv
		i++
	}

	return &InputValues{inputs: inputs}, nil
}

func getAttrInfo(attrName string) (string, data.Type) {
	parts := strings.Split(attrName, "|")

	if len(parts) > 1 {
		dt, _ := data.ToTypeEnum(parts[1])
		return parts[0], dt
	} else {
		return attrName, data.TypeAny
	}
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
		attrs[attr.Name()] = attr.Attribute
	}

	return attrs, nil
}

func (ivs *InputValues) GetDetailedAttrs(scope data.Scope) (map[string]*DetailedAttribute, error) {
	attrs := make(map[string]*DetailedAttribute, len(ivs.inputs))

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
	GetValue(scope data.Scope) (*DetailedAttribute, error)
}

func NewInput(attrName string, attrType data.Type, resolver data.Resolver, value interface{}, newAttr bool) (InputValue, error) {

	if inVal, ok := value.(string); ok && strings.HasPrefix(inVal, "=") {
		//lets assume things that any resolver, assign or expression starts with "="

		strVal := inVal[1:]

		if isFixedResolver(strVal) {
			val, err := data.GetBasicResolver().Resolve(strVal, nil)
			if err != nil {
				return nil, err
			}
			return newFixedInputValue(attrName, attrType, val, newAttr)
		}

		if isExpression(strVal) {
			return newExpressionInputValue(attrName, attrType, strVal, resolver, newAttr)
		}

		if isAssign(strVal) {
			return newAssignInputValue(attrName, attrType, strVal, resolver, newAttr)
		}

		if attrType == data.TypeString {
			//a string is expected, so probably a literal?
			return newFixedInputValue(attrName, attrType, inVal, newAttr)
		}

		return nil, fmt.Errorf("invalid input value '%s' for type %s", inVal, data.TypeString.String())
	} else {
		//not a string, so isn't a mapping
		return newFixedInputValue(attrName, attrType, value, newAttr)
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

func newFixedInputValue(attrName string, attrType data.Type, value interface{}, newAttr bool) (InputValue, error) {

	attr, err := data.NewAttribute(attrName, attrType, value)
	if err != nil {
		return nil, err
	}

	return &fixedInputValue{attr: &DetailedAttribute{Attribute: attr, isNew: newAttr}}, nil
}

type fixedInputValue struct {
	attr *DetailedAttribute
}

func (iv *fixedInputValue) GetValue(scope data.Scope) (*DetailedAttribute, error) {
	return iv.attr, nil
}

func newAssignInputValue(attrName string, attrType data.Type, toResolve string, resolver data.Resolver, newAttr bool) (InputValue, error) {

	return &assignInputValue{attrName: attrName, attrType: attrType, toResolve: toResolve, resolver: resolver, newAttr: newAttr}, nil
}

type assignInputValue struct {
	attrName  string
	attrType  data.Type
	toResolve string
	resolver  data.Resolver
	newAttr   bool
}

func (iv *assignInputValue) GetValue(scope data.Scope) (*DetailedAttribute, error) {

	val, err := resolver.Resolve(iv.toResolve, scope)
	if err != nil {
		return nil, err
	}

	//return data.NewAttribute(iv.attrName, iv.attrType, val)
	attr, err := data.NewAttribute(iv.attrName, iv.attrType, val)
	if err != nil {
		return nil, err
	}

	return &DetailedAttribute{Attribute: attr, isNew: iv.newAttr}, nil
}

func newExpressionInputValue(attrName string, attrType data.Type, exprStr string, resolver data.Resolver, newAttr bool) (InputValue, error) {

	parsedExpr, err := expression.ParseExpression(exprStr)
	if err != nil {
		return nil, err
	}
	return &expressionInputValue{attrName: attrName, attrType: attrType, parsedExpr: parsedExpr, resolver: resolver, newAttr: newAttr}, nil
}

type expressionInputValue struct {
	attrName   string
	attrType   data.Type
	parsedExpr expr.Expr
	resolver   data.Resolver
	newAttr    bool
}

func (iv *expressionInputValue) GetValue(scope data.Scope) (*DetailedAttribute, error) {

	val, err := iv.parsedExpr.EvalWithScope(scope, iv.resolver)
	if err != nil {
		return nil, err
	}

	attr, err := data.NewAttribute(iv.attrName, iv.attrType, val)
	if err != nil {
		return nil, err
	}

	return &DetailedAttribute{Attribute: attr, isNew: iv.newAttr}, nil
}
