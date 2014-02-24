package jsonschema

// NOTE: incomplete.
func normalizeType(dataStruct interface{}) interface{} {
	switch dataStruct.(type) {

	case float32:
		return float64(dataStruct.(float32))

	case int:
		return int64(dataStruct.(int))
	case int8:
		return int64(dataStruct.(int8))
	case int16:
		return int64(dataStruct.(int16))
	case int32:
		return int64(dataStruct.(int32))

	case uint8:
		return int64(dataStruct.(uint8))
	case uint16:
		return int64(dataStruct.(uint16))
	case uint32:
		return int64(dataStruct.(uint32))

	}
	return dataStruct
}
