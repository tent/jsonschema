package jsonschema

func Minimum(s *Schema, dataStruct interface{}) bool {
	actual, ok := dataStruct.(float64)
	if ok == true {
		if actual < *s.Minimum {
			return false
		}
	}
	return true
}
