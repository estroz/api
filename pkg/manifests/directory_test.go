package manifests

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/operator-framework/api/pkg/validation"

	"github.com/ghodss/yaml"
	"github.com/operator-framework/operator-lifecycle-manager/pkg/api/apis/operators/v1alpha1"
)

func TestValidateCSV(t *testing.T) {
	cases := []struct {
		name              string
		inputCSVPath      string
		wantWarn, wantErr bool
	}{
		{
			"CSV with no errors or warnings",
			filepath.Join("testdata", "correct.csv.yaml"), false, false,
		},
		{
			"CSV with a data type mismatch",
			filepath.Join("testdata", "dataTypeMismatch.csv.yaml"), false, true,
		},
	}
	for _, c := range cases {
		b, err := ioutil.ReadFile(c.inputCSVPath)
		if err != nil {
			t.Fatalf("Error reading CSV path %s: %v", c.inputCSVPath, err)
		}
		csv := v1alpha1.ClusterServiceVersion{}
		if err = yaml.Unmarshal(b, &csv); err != nil {
			t.Fatalf("Error unmarshalling CSV at path %s: %v", c.inputCSVPath, err)
		}
		results := validation.DefaultClusterServiceVersionValidators().Apply(&csv)
		if len(results) != 0 {
			if numResults := len(results); numResults != 1 {
				t.Fatalf("Test %s: expected one result, got: %d", c.name, numResults)
			}
			result := results[1]
			if !c.wantErr && !c.wantWarn {
				t.Errorf("Test %s: wanted Errors or Warnings, got: %v", c.name, result)
			}
			if c.wantErr && len(result.Errors) == 0 {
				t.Errorf("Test %s: wanted Error, got: %v", c.name, result)
			}
			if c.wantWarn && len(result.Warnings) == 0 {
				t.Errorf("Test %s: wanted Warning, got: %v", c.name, result)
			}
		} else {
			if c.wantErr {
				t.Errorf("Test %s: wanted Error, got empty return value", c.name)
			}
			if c.wantWarn {
				t.Errorf("Test %s: wanted Warning, got empty return value", c.name)
			}
		}
	}
}
