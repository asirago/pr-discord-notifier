package main

import "testing"

func TestTransformPRbodyIssueLink(t *testing.T) {

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "1 Issue Linked",
			input: "Fixes #237\r\nTested http and https unix sockets both work\r\nnginx work fine with both PR since",
			want:  "Fixes [#237](https://github.com/gotify/server/issues/237)\r\nTested http and https unix sockets both work\r\nnginx work fine with both PR since",
		},
		{
			name:  "No Issue Linked",
			input: "Go embed has been out for a while now and packr recommends moving to //go:embed as well, this PR removes packr from the dependency list and uses go:embed to embed swagger and UI files.",
			want:  "Go embed has been out for a while now and packr recommends moving to //go:embed as well, this PR removes packr from the dependency list and uses go:embed to embed swagger and UI files.",
		},
		{
			name:  "multiple issues linked",
			input: "notifs pushes normally to android\r\ncloses #1337\r\nnotifs pushes normally to ios\r\ncloses #127",
			want:  "notifs pushes normally to android\r\ncloses [#1337](https://github.com/gotify/server/issues/1337)\r\nnotifs pushes normally to ios\r\ncloses [#127](https://github.com/gotify/server/issues/127)",
		},
	}

	url := "https://github.com/gotify/server"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := addIssueURLToPullRequestBody(tt.input, url)

			if tt.want != got {
				t.Errorf("expected: %v, got: %v", tt.want, got)
			}

		})
	}
}
