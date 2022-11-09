package output_test

import (
	"os"
	"testing"

	"github.com/bearer/curio/pkg/commands/process/settings"
	"github.com/bearer/curio/pkg/report/output"
	"github.com/bearer/curio/pkg/report/output/dataflow"
	"github.com/bearer/curio/pkg/report/output/dataflow/types"
	globaltypes "github.com/bearer/curio/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestDataflowDataType(t *testing.T) {
	config := settings.Config{
		CustomDetector: map[string]settings.Rule{
			"logger_leak": {
				Stored: true,
			},
		},
	}

	testCases := []struct {
		Name        string
		FileContent string
		Config      settings.Config
		Want        []types.Datatype
	}{
		{
			Name:        "single detection",
			Config:      config,
			FileContent: `{"type": "schema_classified", "detector_type":"ruby", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.Datatype{
				{
					Name: "Username",
					Detectors: []types.DatatypeDetector{
						{
							Name: "ruby",
							Locations: []types.DatatypeLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
		{
			Name:   "single detection - duplicates",
			Config: config,
			FileContent: `{"type": "schema_classified", "detector_type":"ruby", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}
{"type": "schema_classified", "detector_type":"ruby", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.Datatype{
				{
					Name: "Username",
					Detectors: []types.DatatypeDetector{
						{
							Name: "ruby",
							Locations: []types.DatatypeLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
		{
			Name:   "single detection - with wierd data in report",
			Config: config,
			FileContent: `{"type": "schema_classified", "detector_type":"ruby", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}
{"user": true }`,
			Want: []types.Datatype{
				{
					Name: "Username",
					Detectors: []types.DatatypeDetector{
						{
							Name: "ruby",
							Locations: []types.DatatypeLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
		{
			Name:   "multiple detections - with same object name - deterministic output",
			Config: config,
			FileContent: `{"type": "schema_classified", "detector_type":"ruby", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}
{"type": "schema_classified", "detector_type":"csharp", "source": {"filename": "./users.cs", "line_number": 12}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.Datatype{
				{
					Name: "Username",
					Detectors: []types.DatatypeDetector{
						{
							Name: "csharp",
							Locations: []types.DatatypeLocation{
								{Filename: "./users.cs", LineNumber: 12},
							},
						},
						{
							Name: "ruby",
							Locations: []types.DatatypeLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
		{
			Name:   "multiple detections - with different names - deterministic output",
			Config: config,
			FileContent: `{"type": "schema_classified", "detector_type":"ruby", "source": {"filename": "./users.rb", "line_number": 25}, "value": {"field_name": "User_name", "classification": {"data_type": {"data_category_name": "Username"} ,"decision":{"state": "valid"}}}}
{"type": "schema_classified", "detector_type":"csharp", "source": {"filename": "./users.cs", "line_number": 12}, "value": {"field_name": "address", "classification": {"data_type": {"data_category_name": "Physical Address"} ,"decision":{"state": "valid"}}}}`,
			Want: []types.Datatype{
				{
					Name: "Physical Address",
					Detectors: []types.DatatypeDetector{
						{
							Name: "csharp",
							Locations: []types.DatatypeLocation{
								{Filename: "./users.cs", LineNumber: 12},
							},
						},
					},
				},
				{
					Name: "Username",
					Detectors: []types.DatatypeDetector{
						{
							Name: "ruby",
							Locations: []types.DatatypeLocation{
								{Filename: "./users.rb", LineNumber: 25},
							},
						},
					},
				},
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			file, err := os.CreateTemp("", "*test.jsonlines")
			if err != nil {
				t.Fatalf("failed to create tmp file for report %s", err)
				return
			}
			defer os.Remove(file.Name())
			_, err = file.Write([]byte(test.FileContent))
			if err != nil {
				t.Fatalf("failed to write to tmp file %s", err)
				return
			}
			file.Close()

			detections, err := output.GetDetectorsOutput(globaltypes.Report{
				Path: file.Name(),
			})
			if err != nil {
				t.Fatalf("failed to get detectors output %s", err)
				return
			}

			dataflow, err := dataflow.GetOuput(detections, test.Config)
			if err != nil {
				t.Fatalf("failed to get detectors output %s", err)
				return
			}

			assert.Equal(t, test.Want, dataflow.Datatypes)
		})
	}
}