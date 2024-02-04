package main

import (
	"fmt"
	"regexp"
	"strings"
)

func addIssueURLToPullRequestBody(body, url string) string {

	transformedText := body

	re := regexp.MustCompile(
		`(?i)(close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved) #\d+`,
	)

	for _, v := range re.FindAllString(body, -1) {
		issue := strings.Split(v, "#")[1]

		transformedText = strings.Replace(
			transformedText,
			fmt.Sprintf("#%s", issue),
			fmt.Sprintf("[#%s](%s/issues/%s)", issue, url, issue),
			-1,
		)
	}

	return transformedText
}
