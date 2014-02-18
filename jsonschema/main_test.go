package jsonschema

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var testSuite = flag.String("testsuite", "", "Path to JSON-Schema-Test-Suite.")

func TestDraft4(t *testing.T) {
	testResources := filepath.Join(*testSuite, "tests", "draft4")
	err := filepath.Walk(testResources, processTestFile)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func processTestFile(path string, _ os.FileInfo, err error) error {
	if err != nil {
		return err
	}
	fileStat, err := os.Stat(path)
	if err != nil {
		return err
	}
	if fileStat.IsDir() {
		return nil
	}

	content, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	var testFile []testDescription
	err = json.Unmarshal(content, &testFile)
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
			if len(errorList) > 0 && test.Valid {
				return errors.New("Returned invalid for valid JSON.")
			} else if len(errorList) == 0 && !test.Valid {
				return errors.New("Returned valid for invalid JSON.")
			}
		}
	}
	return nil
}

type testDescription struct {
	Description string
	Schema      json.RawMessage
	Properties  json.RawMessage
	Required    json.RawMessage
	Tests       []testInstance
}

type testInstance struct {
	Description string
	Data        json.RawMessage
	Valid       bool
}
