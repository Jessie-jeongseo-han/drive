package drive_test

import (
	"bytes"
	"encoding/json"
	"testing"

	drive "github.com/odeke-em/drive/src"
)

func TestStructuredRC(t *testing.T) {
	tests := [...]struct {
		rcDir   string
		want    map[string]map[string]interface{}
		wantErr bool
	}{
		0: {
			rcDir: "./testdata/structured",
			want: map[string]map[string]interface{}{
				"global": {"depth": 10, "verbose": false},
				"list":   {"long": true},
				"push":   {"no-prompt": true},
				"pull":   {"depth": 3, "no-prompt": false, "verbose": true},
			},
		},

		1: {
			rcDir: "./testdata/zero-values",
			want: map[string]map[string]interface{}{
				"global": {"force": true, "verbose": true},
				"pull": {
					"desktop-links": false, "force": true, "verbose": false,
					"export": "pdf,doc,rtf,epub", "same-exports-dir": true,
				},
				"open":    {"verbose": true, "file-browser": false, "export": ""},
				"push":    {"trashed": true, "retry-count": 10, "destination": "/tmp"},
				"starred": {"all": true},
			},
		},
	}

	blobify := func(v interface{}) []byte {
		blob, _ := json.Marshal(v)
		return blob
	}

	for i, tt := range tests {
		rcMap, err := drive.ResourceMappings(tt.rcDir)

		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: err=nil", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err=%v", i, err)
			continue
		}

		// Not going to use reflect.DeepEqual because
		// we'll then have trouble comparing any []string.
		gotBlob := blobify(rcMap)
		wantBlob := blobify(tt.want)

		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\n\thave:\n\t%s\n\twant:\n\t%s", i, gotBlob, wantBlob)
		}
	}
}

type cliDefinition struct {
	Depth    *int    `json:"depth"`
	NoPrompt *bool   `json:"no-prompt"`
	Verbose  *bool   `json:"verbose"`
	Long     *bool   `json:"long"`
	Sort     *string `json:"sort"`
}

func TestStructuredRCJSONSifting(t *testing.T) {
	tests := [...]struct {
		rcDir        string
		wantErr      bool
		relevantKeys []string
		want         map[string]interface{}
	}{
		0: {
			rcDir:        "./testdata/structured",
			relevantKeys: []string{"push"},
			want: map[string]interface{}{
				"depth": 10, "verbose": false, "no-prompt": true,
			},
		},

		1: {
			rcDir:        "./testdata/structured",
			relevantKeys: []string{"pull"},
			want: map[string]interface{}{
				"depth": 3, "verbose": true, "no-prompt": false,
			},
		},

		// Since we aren't passing in any keys to get from,
		// make sure we can still read from the global scope
		2: {
			rcDir:        "./testdata/structured",
			relevantKeys: nil,
			want: map[string]interface{}{
				"depth":   10,
				"verbose": false,
			},
		},

		// Ensure that we can read from the
		// old style, non-sectioned .driverc file.
		3: {
			rcDir:        "./testdata/non-sectioned",
			relevantKeys: nil,
			want: map[string]interface{}{
				"depth":     10,
				"verbose":   true,
				"long":      true,
				"no-prompt": false,
				"sort":      "name,date_r",
			},
		},
	}

	for i, tt := range tests {
		filler := cliDefinition{}
		jsonStr, err := drive.JSONStringifySiftedCLITags(filler, tt.rcDir, nil, tt.relevantKeys...)

		if tt.wantErr {
			if err == nil {
				t.Errorf("#%d: err=nil", i)
			}
			continue
		}

		if err != nil {
			t.Errorf("#%d: err=%v", i, err)
			continue
		}

		saveMap := make(map[string]interface{})
		if err := json.Unmarshal([]byte(jsonStr), &saveMap); err != nil {
			t.Errorf("#%d: err=%v", i, err)
			continue
		}

		// Not going to use reflect.DeepEqual because
		// we've have trouble comparing any []string
		gotBlob, _ := json.MarshalIndent(saveMap, "", "  ")
		wantBlob, _ := json.MarshalIndent(tt.want, "", "  ")

		if !bytes.Equal(gotBlob, wantBlob) {
			t.Errorf("#%d:\n\thave: %s\n\twant: %s", i, gotBlob, wantBlob)
		}
	}
}
