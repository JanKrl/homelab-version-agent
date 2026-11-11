package checker

import (
	"strings"
	"testing"
)

func TestParseAptPolicy(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		wantInstalled string
		wantCandidate string
		wantErr       bool
	}{
		{
			name: "Correct and and up to date",
			input: `curl:
  Installed: 8.18.0-1ubuntu2.8
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled: "8.18.0-1ubuntu2.8",
			wantCandidate: "8.18.0-1ubuntu2.8",
			wantErr:       false,
		},
		{
			name: "Correct and outdated",
			input: `curl:
  Installed: 8.18.0-1ubuntu2.7
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled: "8.18.0-1ubuntu2.7",
			wantCandidate: "8.18.0-1ubuntu2.8",
			wantErr:       false,
		},
		{
			name: "Not installed",
			input: `curl:
  Installed: (none)
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled: "(none)",
			wantCandidate: "8.18.0-1ubuntu2.8",
			wantErr:       false,
		},
		{
			name: "Missing Installed",
			input: `curl:
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled: "",
			wantCandidate: "",
			wantErr:       true,
		},
		{
			name: "Missing Candidate",
			input: `curl:
  Installed: 8.18.0-1ubuntu2.7
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled: "",
			wantCandidate: "",
			wantErr:       true,
		},
		{
			name:          "Error message input",
			input:         `E: Unable to locate package curl`,
			wantInstalled: "",
			wantCandidate: "",
			wantErr:       true,
		},
		{
			name:          "Empty input",
			input:         ``,
			wantInstalled: "",
			wantCandidate: "",
			wantErr:       true,
		},
		{
			name:          "Random text input",
			input:         `This is some random text that doesn't conform to the expected format.`,
			wantInstalled: "",
			wantCandidate: "",
			wantErr:       true,
		},
		{
			name: "Multiple packages in input",
			input: `curl:
  Installed: 8.18.0-1ubuntu2.7
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500
wget:
  Installed: 1.20.3-1ubuntu2
  Candidate: 1.20.3-1ubuntu2
  Version table:
 *** 1.20.3-1ubuntu2 500`,
			wantInstalled: "",
			wantCandidate: "",
			wantErr:       true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			installed, candidate, err := parseAptPolicy(strings.NewReader(tc.input))

			if (err != nil) != tc.wantErr {
				t.Errorf("parseAptPolicy() error = %v, wantErr %v", err, tc.wantErr)
				return
			}
			if installed != tc.wantInstalled {
				t.Errorf("parseAptPolicy() gotInstalled = %v, want %v", installed, tc.wantInstalled)
			}
			if candidate != tc.wantCandidate {
				t.Errorf("parseAptPolicy() gotCandidate = %v, want %v", candidate, tc.wantCandidate)
			}
		})
	}
}
