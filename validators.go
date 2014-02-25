package jsonschema

import (
	"fmt"
	"strings"
)

func Minimum(s *Schema, dataStruct interface{}) error {
	var isLarger bool
	switch dataStruct.(type) {
	case int64:
		isLarger = isLargerThanInt(s, dataStruct.(int64))
	case float64:
		isLarger = isLargerThanFloat(s, dataStruct.(float64))
	}
	if isLarger {
		return fmt.Errorf("Value must be larger than %s.", *s.Minimum)
	}
	return nil
}

func isLargerThanInt(s *Schema, data int64) bool {
	if strings.Contains(s.Minimum.String(), ".") {
		flt, err := s.Minimum.Float64()
		if err != nil {
			return false
		}
		if flt > float64(data) {
			return true
		}
	} else {
		intg, err := s.Minimum.Int64()
		if err != nil {
			return false
		}
		if intg > data {
			return true
		}
	}
	return false
}

func isLargerThanFloat(s *Schema, data float64) bool {
	flt, err := s.Minimum.Float64()
	if err != nil {
		return false
	}
	if flt > data {
		return true
	}
	return false
}
