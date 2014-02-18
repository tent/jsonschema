package jsonschema

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestDraft4(t *testing.T) {
	testResources := filepath.Join("JSON-Schema-Test-Suite", "tests", "draft4")
	err := filepath.Walk(testResources, testFileRunner(t))
	if err != nil {
		t.Errorf(err.Error())
	}
}

func testFileRunner(t *testing.T) func(string, os.FileInfo, error) error {
	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Log("Are you sure you ran `git submodule update`?")
			return err
		}
		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		var testFile []testCase
		err = json.NewDecoder(file).Decode(&testFile)
		if err != nil {
			return err
		}

		for _, description := range testFile {
			schema, err := Parse(bytes.NewReader(description.Schema))
			if err != nil {
				return err
			}
			for _, test := range description.Tests {
				errorList := schema.Validate(test.Data)
				message := failureMessage(errorList, test, description, path)
				if len(message) > 0 {
					t.Errorf(message)
				}
			}
		}
		return nil
	}
}

func failureMessage(errorList []ValidationError, tst testInstance, cse testCase, path string) string {
	var validated bool
	if len(errorList) == 0 {
		validated = true
	} else if len(errorList) > 0 {
		validated = false
	}

	var failure string
	if validated && !tst.Valid {
		failure = "schema.Validate validated bad data."
	} else if !validated && tst.Valid {
		failure = "schema.Validate failed to validate good data."
	}

	var message string
	if len(failure) > 0 {
		message = fmt.Sprintf(`%s
file: %s
test case description: %s
schema: %s
test instance description: %s
test data: %s
result of schema.Validate:
	expected: %t
	actual: %t
	actual validation errors: %s

`, failure, path, cse.Description, cse.Schema, tst.Description, tst.Data, tst.Valid, validated, errorList)
	}
	return message
}

type testCase struct {
	Description string          `json:"description"`
	Schema      json.RawMessage `json:"schema"`
	Properties  json.RawMessage `json:"properties"`
	Required    json.RawMessage `json:"required"`
	Tests       []testInstance  `json:"tests"`
}

type testInstance struct {
	Description string          `json:"description"`
	Data        json.RawMessage `json:"data"`
	Valid       bool            `json:"valid"`
}
