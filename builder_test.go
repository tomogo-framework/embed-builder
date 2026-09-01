package embedbuilder

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestBuilderBuild(t *testing.T) {
	t.Parallel()

	timestamp := time.Date(2026, time.August, 31, 12, 34, 56, 123000000, time.FixedZone("test", -5*60*60))
	builder := New().
		SetTitle("Release").
		SetDescription("Ready").
		SetURL("https://example.com/release").
		SetTimestamp(timestamp).
		SetColor(0x57F287).
		SetFooter("Footer", "https://example.com/footer.png").
		SetAuthor("Bot", "https://example.com", "https://example.com/bot.png").
		SetImage("https://example.com/image.png").
		SetThumbnail("https://example.com/thumbnail.png").
		AddField("Status", "Healthy", true)

	embed, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if embed.Title != "Release" || embed.Description != "Ready" {
		t.Fatalf("Build() text = %q, %q", embed.Title, embed.Description)
	}
	if embed.Timestamp != "2026-08-31T17:34:56.123Z" {
		t.Errorf("Build() timestamp = %q", embed.Timestamp)
	}
	if embed.Color == nil || *embed.Color != 0x57F287 {
		t.Errorf("Build() color = %v", embed.Color)
	}
	if len(embed.Fields) != 1 || embed.Fields[0].Name != "Status" {
		t.Errorf("Build() fields = %#v", embed.Fields)
	}
	if embed.Footer == nil || embed.Footer.Text != "Footer" {
		t.Errorf("Build() footer = %#v", embed.Footer)
	}
	if embed.Author == nil || embed.Author.Name != "Bot" {
		t.Errorf("Build() author = %#v", embed.Author)
	}
}

func TestBuildReturnsIndependentSnapshot(t *testing.T) {
	t.Parallel()

	builder := New().
		SetColor(1).
		SetFooter("first").
		SetAuthor("first").
		SetImage("https://example.com/first.png").
		SetThumbnail("attachment://first.png").
		AddField("first", "first", false)

	first, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	*first.Color = 2
	first.Footer.Text = "changed"
	first.Author.Name = "changed"
	first.Image.URL = "changed"
	first.Thumbnail.URL = "changed"
	first.Fields[0].Name = "changed"

	second, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if *second.Color != 1 || second.Footer.Text != "first" || second.Author.Name != "first" {
		t.Fatalf("Build() retained snapshot mutation: %#v", second)
	}
	if second.Image.URL != "https://example.com/first.png" || second.Thumbnail.URL != "attachment://first.png" || second.Fields[0].Name != "first" {
		t.Fatalf("Build() retained snapshot mutation: %#v", second)
	}
}

func TestMetadataSettersReuseBuilderStateWithoutAliasingSnapshots(t *testing.T) {
	t.Parallel()

	builder := New().
		SetColor(1).
		SetFooter("first", "https://example.com/footer.png").
		SetAuthor("first", "https://example.com", "https://example.com/author.png").
		SetImage("https://example.com/first.png").
		SetThumbnail("https://example.com/first-thumbnail.png")

	first, err := builder.Build()
	if err != nil {
		t.Fatalf("first Build() error = %v", err)
	}

	builder.
		SetColor(2).
		SetFooter("second").
		SetAuthor("second").
		SetImage("https://example.com/second.png").
		SetThumbnail("https://example.com/second-thumbnail.png")

	second, err := builder.Build()
	if err != nil {
		t.Fatalf("second Build() error = %v", err)
	}

	if *first.Color != 1 || first.Footer.Text != "first" || first.Author.Name != "first" {
		t.Fatalf("first snapshot changed: %#v", first)
	}
	if *second.Color != 2 || second.Footer.IconURL != "" || second.Author.URL != "" || second.Author.IconURL != "" {
		t.Fatalf("second snapshot retained old metadata: %#v", second)
	}
}

func TestBuilderClearMethods(t *testing.T) {
	t.Parallel()

	builder := New().
		SetTimestamp(time.Now()).
		SetColor(1).
		SetFooter("footer").
		SetAuthor("author").
		SetImage("image").
		SetThumbnail("thumbnail").
		AddField("name", "value", false).
		ClearTimestamp().
		ClearColor().
		ClearFooter().
		ClearAuthor().
		ClearImage().
		ClearThumbnail().
		ClearFields()

	embed, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if embed.Timestamp != "" || embed.Color != nil || embed.Footer != nil || embed.Author != nil {
		t.Fatalf("Build() did not clear metadata: %#v", embed)
	}
	if embed.Image != nil || embed.Thumbnail != nil || len(embed.Fields) != 0 {
		t.Fatalf("Build() did not clear content: %#v", embed)
	}
}

func TestWithCurrentTimestamp(t *testing.T) {
	t.Parallel()

	before := time.Now().UTC()
	embed, err := New().WithCurrentTimestamp().Build()
	after := time.Now().UTC()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	timestamp, err := time.Parse(time.RFC3339Nano, embed.Timestamp)
	if err != nil {
		t.Fatalf("time.Parse() error = %v", err)
	}
	if timestamp.Before(before) || timestamp.After(after) {
		t.Errorf("WithCurrentTimestamp() = %v, want between %v and %v", timestamp, before, after)
	}
}

func TestValidateIndividualTextLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		builder *Builder
		path    string
	}{
		{name: "title", builder: New().SetTitle(strings.Repeat("x", MaxTitleCharacters+1)), path: "title"},
		{name: "description", builder: New().SetDescription(strings.Repeat("x", MaxDescriptionCharacters+1)), path: "description"},
		{name: "field name", builder: New().AddField(strings.Repeat("x", MaxFieldNameCharacters+1), "value", false), path: "fields[0].name"},
		{name: "field value", builder: New().AddField("name", strings.Repeat("x", MaxFieldValueCharacters+1), false), path: "fields[0].value"},
		{name: "footer", builder: New().SetFooter(strings.Repeat("x", MaxFooterCharacters+1)), path: "footer.text"},
		{name: "author", builder: New().SetAuthor(strings.Repeat("x", MaxAuthorNameCharacters+1)), path: "author.name"},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			validationError := requireValidationError(t, test.builder.Validate())
			if !containsPath(validationError, test.path) {
				t.Errorf("Validate() violations = %#v, want path %q", validationError.Violations, test.path)
			}
		})
	}
}

func TestValidateAcceptsInclusiveLimits(t *testing.T) {
	t.Parallel()

	tests := []*Builder{
		New().SetTitle(strings.Repeat("x", MaxTitleCharacters)),
		New().SetDescription(strings.Repeat("x", MaxDescriptionCharacters)),
		New().AddField(strings.Repeat("x", MaxFieldNameCharacters), "value", false),
		New().AddField("name", strings.Repeat("x", MaxFieldValueCharacters), false),
		New().SetFooter(strings.Repeat("x", MaxFooterCharacters)),
		New().SetAuthor(strings.Repeat("x", MaxAuthorNameCharacters)),
	}

	for index, builder := range tests {
		if err := builder.Validate(); err != nil {
			t.Errorf("test %d: Validate() error = %v", index, err)
		}
	}
}

func TestValidateCountsUnicodeCharactersAndTrimsWhitespace(t *testing.T) {
	t.Parallel()

	valid := New().SetTitle(" \t" + strings.Repeat("🙂", MaxTitleCharacters) + "\n ")
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() Unicode boundary error = %v", err)
	}

	invalid := New().SetTitle(strings.Repeat("🙂", MaxTitleCharacters+1))
	requireValidationError(t, invalid.Validate())
}

func TestValidateFieldAndColorLimits(t *testing.T) {
	t.Parallel()

	builder := New().SetColor(MaxColor + 1)
	for index := 0; index <= MaxFields; index++ {
		builder.AddField("name", "value", false)
	}

	validationError := requireValidationError(t, builder.Validate())
	if !containsPath(validationError, "fields") || !containsPath(validationError, "color") {
		t.Fatalf("Validate() violations = %#v", validationError.Violations)
	}
}

func TestValidateTotalCharacters(t *testing.T) {
	t.Parallel()

	valid := New().
		SetDescription(strings.Repeat("d", MaxDescriptionCharacters)).
		SetFooter(strings.Repeat("f", MaxTotalCharacters-MaxDescriptionCharacters))
	if err := valid.Validate(); err != nil {
		t.Fatalf("Validate() exact total error = %v", err)
	}

	invalid := valid.SetTitle("x")
	validationError := requireValidationError(t, invalid.Validate())
	if !containsPath(validationError, "total_characters") {
		t.Errorf("Validate() violations = %#v", validationError.Violations)
	}

	if embed, err := invalid.Build(); err == nil || !reflect.DeepEqual(embed, Embed{}) {
		t.Fatalf("Build() = %#v, %v; want zero embed and validation error", embed, err)
	}
}

func TestValidateEmbedsSharedLimits(t *testing.T) {
	t.Parallel()

	first := Embed{Description: strings.Repeat("a", 3001)}
	second := Embed{Description: strings.Repeat("b", 3000)}
	validationError := requireValidationError(t, ValidateEmbeds(first, second))
	if !containsPath(validationError, "embeds.total_characters") {
		t.Errorf("ValidateEmbeds() violations = %#v", validationError.Violations)
	}

	embeds := make([]Embed, MaxEmbedsPerMessage+1)
	validationError = requireValidationError(t, ValidateEmbeds(embeds...))
	if !containsPath(validationError, "embeds") {
		t.Errorf("ValidateEmbeds() violations = %#v", validationError.Violations)
	}
}

func TestValidationCollectsAllViolations(t *testing.T) {
	t.Parallel()

	builder := New().
		SetTitle(strings.Repeat("t", MaxTitleCharacters+1)).
		SetDescription(strings.Repeat("d", MaxDescriptionCharacters+1)).
		SetFooter(strings.Repeat("f", MaxFooterCharacters+1)).
		SetAuthor(strings.Repeat("a", MaxAuthorNameCharacters+1))

	validationError := requireValidationError(t, builder.Validate())
	if len(validationError.Violations) != 5 {
		t.Fatalf("Validate() violations = %#v, want 5", validationError.Violations)
	}
}

func TestBuilderCachedValidationMatchesEmbedValidation(t *testing.T) {
	t.Parallel()

	builders := []*Builder{
		New(),
		New().
			SetTitle(strings.Repeat("t", MaxTitleCharacters+1)).
			SetTitle("valid").
			SetDescription("description").
			SetFooter("footer").
			SetAuthor("author"),
		New().
			SetTitle(strings.Repeat("t", MaxTitleCharacters+1)).
			SetDescription(strings.Repeat("d", MaxDescriptionCharacters+1)).
			SetFooter(strings.Repeat("f", MaxFooterCharacters+1)).
			SetAuthor(strings.Repeat("a", MaxAuthorNameCharacters+1)).
			AddFields(
				Field{Name: strings.Repeat("n", MaxFieldNameCharacters+1), Value: "value"},
				Field{Name: "name", Value: strings.Repeat("v", MaxFieldValueCharacters+1)},
			),
		New().
			AddField(strings.Repeat("n", MaxFieldNameCharacters+1), "value", false).
			ClearFields().
			AddField("name", "value", false).
			ClearFooter().
			ClearAuthor(),
		New().
			SetURL("not a URL").
			SetImage("not an image URL"),
	}

	for index, builder := range builders {
		cachedError := builder.Validate()
		directError := ValidateEmbed(builder.embed)
		if !reflect.DeepEqual(cachedError, directError) {
			t.Errorf("builder %d: cached error = %#v, direct error = %#v", index, cachedError, directError)
		}
	}
}

func TestNilBuilder(t *testing.T) {
	t.Parallel()

	var builder *Builder
	if !errors.Is(builder.Validate(), ErrNilBuilder) {
		t.Errorf("Validate() error = %v", builder.Validate())
	}
	if _, err := builder.Build(); !errors.Is(err, ErrNilBuilder) {
		t.Errorf("Build() error = %v", err)
	}
}

func TestJSONShape(t *testing.T) {
	t.Parallel()

	embed, err := New().SetColor(0).AddField("name", "value", true).Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	payload, err := json.Marshal(embed)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	want := `{"color":0,"fields":[{"name":"name","value":"value","inline":true}]}`
	if string(payload) != want {
		t.Errorf("json.Marshal() = %s, want %s", payload, want)
	}
}

func requireValidationError(t *testing.T, err error) *ValidationError {
	t.Helper()

	var validationError *ValidationError
	if !errors.As(err, &validationError) {
		t.Fatalf("error = %v, want *ValidationError", err)
	}

	return validationError
}

func containsPath(validationError *ValidationError, path string) bool {
	for _, violation := range validationError.Violations {
		if violation.Path == path {
			return true
		}
	}

	return false
}
