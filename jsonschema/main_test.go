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
		schema, err := Parse(bytes.NewReader(description.schema))
		if err != nil {
			return err
		}
		for _, test := range description.tests {
			errorList := schema.Validate(test.data)
			if len(errorList) > 0 && test.valid {
				return errors.New("Returned invalid for valid JSON.")
			} else if len(errorList) == 0 && !test.valid {
				return errors.New("Returned valid for invalid JSON.")
			}
		}
	}
	return nil
}

type testDescription struct {
	description string
	schema      json.RawMessage
	properties  json.RawMessage
	required    json.RawMessage
	tests       []testInstance
}

type testInstance struct {
	description string
	data        json.RawMessage
	valid       bool
}
