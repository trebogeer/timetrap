package util

func AssertString(i interface{}, def string) string {
	switch i.(type) {
	case string:
		return i.(string)
	default:
		return def
	}
}

func AssertFloat64(i interface{}, def float64) float64 {
	switch i.(type) {
	case int:
		return float64(i.(int))
	case float64:
		return i.(float64)
	case int64:
		return float64(i.(int64))
	case int32:
		return float64(i.(int32))
	case float32:
		return float64(i.(float32))
	default:
		return def
	}
}


// TODO expand
func AssertInt64(i interface{}, def int64) int64 {
    switch i.(type) {
    case int:
        return int64(i.(int))
    case int64:
        return i.(int64)
    default:
        return def
    }
}
