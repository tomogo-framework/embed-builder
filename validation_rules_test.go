package embedbuilder

import (
	"strings"
	"testing"
)

func TestStrictValidationAcceptsDiscordURLsAndTimestamp(t *testing.T) {
	t.Parallel()

	embed := Embed{
		URL:       "https://example.com/embed",
		Timestamp: "2026-08-31T12:34:56.123Z",
		Footer:    &Footer{Text: "footer", IconURL: "attachment://footer.png"},
		Image:     &Media{URL: "https://example.com/image.png"},
		Thumbnail: &Media{URL: "attachment://thumbnail.png"},
		Author: &Author{
			Name:    "author",
			URL:     "http://example.com/author",
			IconURL: "https://example.com/author.png",
		},
		Fields: []Field{{Name: "name", Value: "value"}},
	}

	if err := ValidateEmbed(embed); err != nil {
		t.Fatalf("ValidateEmbed() error = %v", err)
	}
}

func TestStrictValidationRules(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		embed Embed
		path  string
		rule  string
	}{
		{name: "embed URL", embed: Embed{URL: "ftp://example.com"}, path: "url", rule: ruleHTTPURL},
		{name: "timestamp", embed: Embed{Timestamp: "yesterday"}, path: "timestamp", rule: ruleTimestamp},
		{name: "footer text", embed: Embed{Footer: &Footer{}}, path: "footer.text", rule: ruleRequired},
		{name: "footer icon", embed: Embed{Footer: &Footer{Text: "text", IconURL: "ftp://example.com"}}, path: "footer.icon_url", rule: ruleMediaURL},
		{name: "author name", embed: Embed{Author: &Author{Name: "  "}}, path: "author.name", rule: ruleRequired},
		{name: "author URL", embed: Embed{Author: &Author{Name: "name", URL: "attachment://author"}}, path: "author.url", rule: ruleHTTPURL},
		{name: "author icon", embed: Embed{Author: &Author{Name: "name", IconURL: "file.png"}}, path: "author.icon_url", rule: ruleMediaURL},
		{name: "image empty", embed: Embed{Image: &Media{}}, path: "image.url", rule: ruleMediaURL},
		{name: "attachment path", embed: Embed{Image: &Media{URL: "attachment://folder/image.png"}}, path: "image.url", rule: ruleMediaURL},
		{name: "thumbnail scheme", embed: Embed{Thumbnail: &Media{URL: "ftp://example.com/image.png"}}, path: "thumbnail.url", rule: ruleMediaURL},
		{name: "field name", embed: Embed{Fields: []Field{{Value: "value"}}}, path: "fields[0].name", rule: ruleRequired},
		{name: "field value", embed: Embed{Fields: []Field{{Name: "name", Value: "\t"}}}, path: "fields[0].value", rule: ruleRequired},
		{name: "UTF-8", embed: Embed{Title: string([]byte{0xff})}, path: "title", rule: ruleUTF8},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			validationError := requireValidationError(t, ValidateEmbed(test.embed))
			if !containsRule(validationError, test.path, test.rule) {
				t.Errorf("ValidateEmbed() violations = %#v", validationError.Violations)
			}
			if !strings.Contains(validationError.Error(), test.rule) {
				t.Errorf("ValidationError.Error() = %q", validationError.Error())
			}
		})
	}
}

func containsRule(validationError *ValidationError, path, rule string) bool {
	for _, violation := range validationError.Violations {
		if violation.Path == path && violation.Rule == rule {
			return true
		}
	}

	return false
}
