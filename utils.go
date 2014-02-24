package jsonschema

// Will also convert int32 to int64, float32 to float64, etc.
func typeSwitch(dataStruct interface{}) (interface{}, string) {
	switch dataStruct.(type) {
	case int64:
		return dataStruct, "int64"
	case float64:
		return dataStruct, "float64"
	case map[string]interface{}:
		return dataStruct, "map[string]interface{}"
	default:
		return dataStruct, ""
	}
}
