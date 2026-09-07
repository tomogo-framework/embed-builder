package embedbuilder

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ErrNilBuilder is returned when Build or Validate is called on a nil Builder.
var ErrNilBuilder = errors.New("embedbuilder: nil builder")

// ErrNilCollection is returned when a collection method requiring state is called on a nil CollectionBuilder.
var ErrNilCollection = errors.New("embedbuilder: nil collection")

// Violation describes one Discord embed validation violation. Rule is populated for structural violations and values
// too large to represent as int; Actual and Limit are populated for representable size limits.
type Violation struct {
	Path   string
	Rule   string
	Actual int
	Limit  int
}

// ValidationError contains every limit violation found in one validation pass.
type ValidationError struct {
	Violations []Violation
}

// Error implements error.
func (e *ValidationError) Error() string {
	if e == nil || len(e.Violations) == 0 {
		return "embedbuilder: validation failed"
	}

	var message strings.Builder
	message.WriteString("embedbuilder: validation failed: ")
	for index, violation := range e.Violations {
		if index != 0 {
			message.WriteString("; ")
		}
		if violation.Rule != "" {
			message.WriteString(violation.Path)
			message.WriteByte(' ')
			message.WriteString(violation.Rule)

			continue
		}

		fmt.Fprintf(&message, "%s has %d, limit is %d", violation.Path, violation.Actual, violation.Limit)
	}

	return message.String()
}

// ValidateEmbed validates a standalone embed against Discord's per-embed limits. Use ValidateEmbeds when multiple
// embeds will share one message.
func ValidateEmbed(embed Embed) error {
	violations, _ := inspectEmbed(embed, -1, true)

	return validationResult(violations)
}

// ValidateEmbeds validates a complete message's embeds, including Discord's 10-embed limit and the shared
// 6,000-character limit.
func ValidateEmbeds(embeds ...Embed) error {
	violations := make([]Violation, 0)
	if len(embeds) > MaxEmbedsPerMessage {
		violations = append(violations, Violation{
			Path:   "embeds",
			Actual: len(embeds),
			Limit:  MaxEmbedsPerMessage,
		})
	}

	total := 0
	for index, embed := range embeds {
		embedViolations, characters := inspectEmbed(embed, index, false)
		violations = append(violations, embedViolations...)
		total += characters
	}
	if total > MaxTotalCharacters {
		violations = append(violations, Violation{
			Path:   "embeds.total_characters",
			Actual: total,
			Limit:  MaxTotalCharacters,
		})
	}

	return validationResult(violations)
}

func validateBuilder(builder *Builder) error {
	violations := make([]Violation, 0)
	violations = appendViolation(violations, "title", builder.counts.title, MaxTitleCharacters)
	violations = appendViolation(violations, "description", builder.counts.description, MaxDescriptionCharacters)
	violations = appendViolation(violations, "fields", len(builder.embed.Fields), MaxFields)
	violations = append(violations, builder.counts.fieldViolations...)
	violations = appendViolation(violations, "footer.text", builder.counts.footer, MaxFooterCharacters)
	violations = appendViolation(violations, "author.name", builder.counts.author, MaxAuthorNameCharacters)
	if builder.embed.Color != nil {
		violations = appendColorViolation(violations, "color", *builder.embed.Color)
	}

	violations = appendViolation(violations, "total_characters", builder.counts.total, MaxTotalCharacters)
	if builder.counts.rulesDirty {
		builder.counts.ruleViolations = inspectEmbedRules(builder.embed, -1)
		builder.counts.rulesDirty = false
	}

	violations = append(violations, builder.counts.ruleViolations...)

	return validationResult(violations)
}

func inspectEmbed(embed Embed, embedIndex int, validateTotal bool) ([]Violation, int) {
	violations := make([]Violation, 0)
	total := 0

	violations, total = inspectText(violations, total, embedIndex, "title", embed.Title, MaxTitleCharacters)
	violations, total = inspectText(violations, total, embedIndex, "description", embed.Description, MaxDescriptionCharacters)

	if len(embed.Fields) > MaxFields {
		violations = append(violations, Violation{
			Path:   embedViolationPath(embedIndex, "fields"),
			Actual: len(embed.Fields),
			Limit:  MaxFields,
		})
	}
	for index, field := range embed.Fields {
		nameCharacters := characterCount(field.Name)
		total += nameCharacters
		if nameCharacters > MaxFieldNameCharacters {
			violations = append(violations, Violation{
				Path:   embedViolationPath(embedIndex, fieldViolationPath(index, "name")),
				Actual: nameCharacters,
				Limit:  MaxFieldNameCharacters,
			})
		}

		valueCharacters := characterCount(field.Value)
		total += valueCharacters
		if valueCharacters > MaxFieldValueCharacters {
			violations = append(violations, Violation{
				Path:   embedViolationPath(embedIndex, fieldViolationPath(index, "value")),
				Actual: valueCharacters,
				Limit:  MaxFieldValueCharacters,
			})
		}
	}

	if embed.Footer != nil {
		violations, total = inspectText(violations, total, embedIndex, "footer.text", embed.Footer.Text, MaxFooterCharacters)
	}
	if embed.Author != nil {
		violations, total = inspectText(violations, total, embedIndex, "author.name", embed.Author.Name, MaxAuthorNameCharacters)
	}
	if embed.Color != nil && *embed.Color > MaxColor {
		violations = appendColorViolation(violations, embedViolationPath(embedIndex, "color"), *embed.Color)
	}
	if validateTotal && total > MaxTotalCharacters {
		violations = append(violations, Violation{
			Path:   embedViolationPath(embedIndex, "total_characters"),
			Actual: total,
			Limit:  MaxTotalCharacters,
		})
	}

	violations = append(violations, inspectEmbedRules(embed, embedIndex)...)

	return violations, total
}

func inspectText(violations []Violation, total, embedIndex int, property, value string, limit int) ([]Violation, int) {
	characters := characterCount(value)
	total += characters
	if characters > limit {
		violations = append(violations, Violation{
			Path:   embedViolationPath(embedIndex, property),
			Actual: characters,
			Limit:  limit,
		})
	}

	return violations, total
}

func characterCount(value string) int {
	trimmed := strings.TrimSpace(value)
	index := 0
	for len(trimmed)-index >= 8 {
		word := binary.LittleEndian.Uint64([]byte(trimmed[index : index+8]))
		if word&0x8080808080808080 != 0 {
			return index + utf8.RuneCountInString(trimmed[index:])
		}

		index += 8
	}

	for ; index < len(trimmed); index++ {
		if trimmed[index] >= utf8.RuneSelf {
			return index + utf8.RuneCountInString(trimmed[index:])
		}
	}

	return len(trimmed)
}

func appendColorViolation(violations []Violation, path string, color uint32) []Violation {
	if color <= MaxColor {
		return violations
	}

	violation := Violation{
		Path: path,
	}
	if strconv.IntSize == 32 && color > 1<<31-1 {
		violation.Rule = fmt.Sprintf("has %d, limit is %d", color, MaxColor)
	} else {
		violation.Actual = int(color)
		violation.Limit = MaxColor
	}

	return append(violations, violation)
}

func appendViolation(violations []Violation, path string, actual, limit int) []Violation {
	if actual <= limit {
		return violations
	}

	return append(violations, Violation{
		Path:   path,
		Actual: actual,
		Limit:  limit,
	})
}

func fieldViolationPath(index int, property string) string {
	return "fields[" + strconv.Itoa(index) + "]." + property
}

func embedViolationPath(index int, property string) string {
	if index < 0 {
		return property
	}

	return "embeds[" + strconv.Itoa(index) + "]." + property
}

func validationResult(violations []Violation) error {
	if len(violations) == 0 {
		return nil
	}

	return &ValidationError{Violations: violations}
}
