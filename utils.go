package jsonschema

import (
	"fmt"
	"reflect"
)

func normalizeType(dataStruct interface{}) (interface{}, error) {
	var data interface{}
	var err error
	switch dataStruct.(type) {

	case float32:
		data = float64(dataStruct.(float32))
	case float64:
		data = dataStruct

	case int:
		data = int64(dataStruct.(int))
	case int8:
		data = int64(dataStruct.(int8))
	case int16:
		data = int64(dataStruct.(int16))
	case int32:
		data = int64(dataStruct.(int32))
	case int64:
		data = dataStruct

	case uint8:
		data = int64(dataStruct.(uint8))
	case uint16:
		data = int64(dataStruct.(uint16))
	case uint32:
		data = int64(dataStruct.(uint32))

	case bool:
		data = dataStruct
	case nil:
		data = dataStruct
	case string:
		data = dataStruct
	case []interface{}:
		data = dataStruct
	case map[string]interface{}:
		data = dataStruct

	default:
		err = fmt.Errorf("%s is not a supported type.", reflect.TypeOf(dataStruct))
	}

	if err != nil {
		return dataStruct, err
	}
	return data, nil
}
