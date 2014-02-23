package jsonschema

import (
	"fmt"
	"strings"
)

func IntMinimum(s *Schema, data int64) error {
	if strings.Contains(s.Minimum.String(), ".") {
		flt, err := s.Minimum.Float64()
		if err != nil {
			return nil
		}
		if float64(data) < flt {
			return fmt.Errorf("Value must be larger than %s.", *s.Minimum)
		}
	} else {
		intg, err := s.Minimum.Int64()
		if err != nil {
			return nil
		}
		if data < intg {
			return fmt.Errorf("Value must be larger than %s.", *s.Minimum)
		}
	}
	return nil
}

func FloatMinimum(s *Schema, data float64) error {
	flt, err := s.Minimum.Float64()
	if err != nil {
		return nil
	}
	if data < flt {
		return fmt.Errorf("Value must be larger than %s.", *s.Minimum)
	}
	return nil
}
