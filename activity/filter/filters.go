package filter

type NonZeroFilter struct {
}

type ConditionalFilter struct {
}

func (*NonZeroFilter) FilterOut(val interface{}, condition bool) bool {
	return !IsNonZero(val)
}

func IsNonZero(val interface{}) bool {

	switch t := val.(type) {
	case int:
		return t != 0
	case float64:
		return t != 0.0
	case []int:
		for _, v := range t {
			if v != 0 {
				return true
			}
		}
		return false
	case []float64:
		for _, v := range t {
			if v != 0.0 {
				return true
			}
		}
		return false
	}

	//todo handle unsupported type
	return true
}

func (*ConditionalFilter) FilterOut(val interface{}, condition bool) bool {
	return condition
}
