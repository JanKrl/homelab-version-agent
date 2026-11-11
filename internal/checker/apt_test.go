package checker

import (
	"strings"
	"testing"
)

func TestParseAptPolicy(t *testing.T) {
	tests := []struct {
		name            string
		input           string
		wantInstalled   string
		wantestandidate string
		wantErr         bool
	}{
		{
			name: "Correct and and up to date",
			input: `curl:
  Installed: 8.18.0-1ubuntu2.8
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled:   "8.18.0-1ubuntu2.8",
			wantestandidate: "8.18.0-1ubuntu2.8",
			wantErr:         false,
		},
		{
			name: "Correct and outdated",
			input: `curl:
  Installed: 8.18.0-1ubuntu2.7
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled:   "8.18.0-1ubuntu2.7",
			wantestandidate: "8.18.0-1ubuntu2.8",
			wantErr:         false,
		},
		{
			name: "Not installed",
			input: `curl:
  Installed: (none)
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled:   "(none)",
			wantestandidate: "8.18.0-1ubuntu2.8",
			wantErr:         false,
		},
		{
			name: "Missing Installed",
			input: `curl:
  Candidate: 8.18.0-1ubuntu2.8
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled:   "",
			wantestandidate: "",
			wantErr:         true,
		},
		{
			name: "Missing Candidate",
			input: `curl:
  Installed: 8.18.0-1ubuntu2.7
  Version table:
 *** 8.18.0-1ubuntu2.8 500`,
			wantInstalled:   "",
			wantestandidate: "",
			wantErr:         true,
		},
		{
			name:            "Error message input",
			input:           `E: Unable to locate package curl`,
			wantInstalled:   "",
			wantestandidate: "",
			wantErr:         true,
		},
		{
			name:            "Empty input",
			input:           ``,
			wantInstalled:   "",
			wantestandidate: "",
			wantErr:         true,
		},
		{
			name:            "Random text input",
			input:           `This is some random text that doesn't conform to the expected format.`,
			wantInstalled:   "",
			wantestandidate: "",
			wantErr:         true,
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
			wantInstalled:   "",
			wantestandidate: "",
			wantErr:         true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			installed, candidate, err := parseAptPolicy(strings.NewReader(test.input))

			if (err != nil) != test.wantErr {
				t.Errorf("parseAptPolicy() error = %v, wantErr %v", err, test.wantErr)
				return
			}
			if installed != test.wantInstalled {
				t.Errorf("parseAptPolicy() gotInstalled = %v, want %v", installed, test.wantInstalled)
			}
			if candidate != test.wantestandidate {
				t.Errorf("parseAptPolicy() gotestandidate = %v, want %v", candidate, test.wantestandidate)
			}
		})
	}
}
