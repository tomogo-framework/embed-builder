package embedbuilder

import (
	"net/url"
	"strings"
	"time"
	"unicode/utf8"
)

const (
	ruleRequired  = "must not be empty"
	ruleUTF8      = "must contain valid UTF-8"
	ruleHTTPURL   = "must be an absolute HTTP(S) URL"
	ruleMediaURL  = "must be an absolute HTTP(S) or attachment URL"
	ruleTimestamp = "must be an ISO 8601 timestamp"
)

func inspectEmbedRules(embed Embed, embedIndex int) []Violation {
	violations := make([]Violation, 0)
	violations = inspectOptionalText(violations, embedIndex, "title", embed.Title)
	violations = inspectOptionalText(violations, embedIndex, "description", embed.Description)
	violations = inspectOptionalHTTPURL(violations, embedIndex, "url", embed.URL)
	violations = inspectOptionalTimestamp(violations, embedIndex, "timestamp", embed.Timestamp)

	for index, field := range embed.Fields {
		violations = inspectRequiredFieldText(violations, embedIndex, index, "name", field.Name)
		violations = inspectRequiredFieldText(violations, embedIndex, index, "value", field.Value)
	}

	if embed.Footer != nil {
		violations = inspectRequiredText(violations, embedIndex, "footer.text", embed.Footer.Text)
		violations = inspectOptionalMediaURL(violations, embedIndex, "footer.icon_url", embed.Footer.IconURL)
	}
	if embed.Author != nil {
		violations = inspectRequiredText(violations, embedIndex, "author.name", embed.Author.Name)
		violations = inspectOptionalHTTPURL(violations, embedIndex, "author.url", embed.Author.URL)
		violations = inspectOptionalMediaURL(violations, embedIndex, "author.icon_url", embed.Author.IconURL)
	}
	if embed.Image != nil {
		violations = inspectRequiredMediaURL(violations, embedIndex, "image.url", embed.Image.URL)
	}
	if embed.Thumbnail != nil {
		violations = inspectRequiredMediaURL(violations, embedIndex, "thumbnail.url", embed.Thumbnail.URL)
	}

	return violations
}

func inspectOptionalText(violations []Violation, embedIndex int, path, value string) []Violation {
	if value == "" {
		return violations
	}

	return inspectUTF8(violations, embedIndex, path, value)
}

func inspectRequiredText(violations []Violation, embedIndex int, path, value string) []Violation {
	if !utf8.ValidString(value) {
		return appendRuleViolation(violations, embedIndex, path, ruleUTF8)
	}
	if strings.TrimSpace(value) == "" {
		return appendRuleViolation(violations, embedIndex, path, ruleRequired)
	}

	return violations
}

func inspectRequiredFieldText(violations []Violation, embedIndex, fieldIndex int, property, value string) []Violation {
	if !utf8.ValidString(value) {
		return appendRuleViolation(violations, embedIndex, fieldViolationPath(fieldIndex, property), ruleUTF8)
	}
	if strings.TrimSpace(value) == "" {
		return appendRuleViolation(violations, embedIndex, fieldViolationPath(fieldIndex, property), ruleRequired)
	}

	return violations
}

func inspectUTF8(violations []Violation, embedIndex int, path, value string) []Violation {
	if utf8.ValidString(value) {
		return violations
	}

	return appendRuleViolation(violations, embedIndex, path, ruleUTF8)
}

func inspectOptionalHTTPURL(violations []Violation, embedIndex int, path, value string) []Violation {
	if value == "" {
		return violations
	}
	if !utf8.ValidString(value) {
		return appendRuleViolation(violations, embedIndex, path, ruleUTF8)
	}
	if !validHTTPURL(value) {
		return appendRuleViolation(violations, embedIndex, path, ruleHTTPURL)
	}

	return violations
}

func inspectOptionalMediaURL(violations []Violation, embedIndex int, path, value string) []Violation {
	if value == "" {
		return violations
	}

	return inspectRequiredMediaURL(violations, embedIndex, path, value)
}

func inspectRequiredMediaURL(violations []Violation, embedIndex int, path, value string) []Violation {
	if !utf8.ValidString(value) {
		return appendRuleViolation(violations, embedIndex, path, ruleUTF8)
	}
	if !validMediaURL(value) {
		return appendRuleViolation(violations, embedIndex, path, ruleMediaURL)
	}

	return violations
}

func inspectOptionalTimestamp(violations []Violation, embedIndex int, path, value string) []Violation {
	if value == "" {
		return violations
	}
	if !utf8.ValidString(value) {
		return appendRuleViolation(violations, embedIndex, path, ruleUTF8)
	}
	if _, err := time.Parse(time.RFC3339Nano, value); err != nil {
		return appendRuleViolation(violations, embedIndex, path, ruleTimestamp)
	}

	return violations
}

func validHTTPURL(value string) bool {
	if !utf8.ValidString(value) {
		return false
	}

	parsed, err := url.Parse(value)
	if err != nil || parsed.Host == "" {
		return false
	}

	return strings.EqualFold(parsed.Scheme, "http") || strings.EqualFold(parsed.Scheme, "https")
}

func validMediaURL(value string) bool {
	if validHTTPURL(value) {
		return true
	}
	if !utf8.ValidString(value) || !strings.HasPrefix(value, "attachment://") {
		return false
	}

	filename := strings.TrimPrefix(value, "attachment://")

	return filename != "" && !strings.ContainsAny(filename, `/\`)
}

func appendRuleViolation(violations []Violation, embedIndex int, path, rule string) []Violation {
	return append(violations, Violation{
		Path: embedViolationPath(embedIndex, path),
		Rule: rule,
	})
}
