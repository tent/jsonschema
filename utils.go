package jsonschema

func typeSwitch(dataStruct interface{}) string {
	switch dataStruct.(type) {
	case int64:
		return "int64"
	case float64:
		return "float64"
	default:
		return ""
	}
}
