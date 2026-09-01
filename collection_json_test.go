package embedbuilder

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCollectionBuilder(t *testing.T) {
	t.Parallel()

	source := Embed{Title: "first", Fields: []Field{{Name: "name", Value: "value"}}}
	collection := NewCollection().Add(source)
	if err := collection.AddBuilder(New().SetTitle("second")); err != nil {
		t.Fatalf("AddBuilder() error = %v", err)
	}

	source.Title = "changed"
	source.Fields[0].Name = "changed"
	embeds, err := collection.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if collection.Count() != 2 || embeds[0].Title != "first" || embeds[0].Fields[0].Name != "name" {
		t.Fatalf("Build() embeds = %#v", embeds)
	}

	embeds[0].Fields[0].Name = "output mutation"
	again, err := collection.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if again[0].Fields[0].Name != "name" {
		t.Error("Build() result aliases collection state")
	}

	collection.Clear()
	if collection.Count() != 0 {
		t.Errorf("Count() after Clear = %d", collection.Count())
	}
}

func TestCollectionSharedValidation(t *testing.T) {
	t.Parallel()

	collection := NewCollection().
		Add(Embed{Description: strings.Repeat("a", 3001)}).
		Add(Embed{Description: strings.Repeat("b", 3000)})
	validationError := requireValidationError(t, collection.Validate())
	if !containsPath(validationError, "embeds.total_characters") {
		t.Errorf("Validate() violations = %#v", validationError.Violations)
	}

	var nilCollection *CollectionBuilder
	if !errors.Is(nilCollection.Validate(), ErrNilCollection) {
		t.Errorf("nil Validate() error = %v", nilCollection.Validate())
	}
	if _, err := nilCollection.Build(); !errors.Is(err, ErrNilCollection) {
		t.Errorf("nil Build() error = %v", err)
	}
	if !errors.Is(nilCollection.AddBuilder(New()), ErrNilCollection) {
		t.Errorf("nil AddBuilder() error = %v", nilCollection.AddBuilder(New()))
	}
}

func TestBuilderJSON(t *testing.T) {
	t.Parallel()

	builder := New().SetTitle("json").SetColor(ColorBlurple).AddField("name", "value", true)
	embed, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	want, err := json.Marshal(embed)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	payload, err := builder.BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON() error = %v", err)
	}
	if !reflect.DeepEqual(payload, want) {
		t.Errorf("BuildJSON() = %s, want %s", payload, want)
	}

	prefix := []byte("prefix:")
	appended, err := builder.AppendJSON(append([]byte(nil), prefix...))
	if err != nil {
		t.Fatalf("AppendJSON() error = %v", err)
	}
	if !reflect.DeepEqual(appended, append(prefix, want...)) {
		t.Errorf("AppendJSON() = %s", appended)
	}

	invalid := New().SetImage("invalid")
	destination := []byte("unchanged")
	result, err := invalid.AppendJSON(destination)
	if err == nil || !reflect.DeepEqual(result, destination) {
		t.Errorf("invalid AppendJSON() = %q, %v", result, err)
	}
}

func TestBuilderJSONWithOptionalMetadata(t *testing.T) {
	t.Parallel()

	builder := New().
		SetTitle("metadata").
		SetURL("https://example.com/embed").
		SetTimestamp(time.Date(2026, time.August, 31, 12, 34, 56, 0, time.UTC)).
		SetFooter("footer", "https://example.com/footer.png").
		SetAuthor("author", "https://example.com/author", "https://example.com/author.png").
		SetImage("https://example.com/image.png").
		SetThumbnail("https://example.com/thumbnail.png")

	embed, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	want, err := json.Marshal(embed)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	payload, err := builder.BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON() error = %v", err)
	}
	if !reflect.DeepEqual(payload, want) {
		t.Errorf("BuildJSON() = %s, want %s", payload, want)
	}
}

func TestCollectionJSON(t *testing.T) {
	t.Parallel()

	collection := NewCollection().Add(Embed{Title: "one"}).Add(Embed{Title: "two"})
	embeds, err := collection.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	want, err := json.Marshal(embeds)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	payload, err := collection.BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON() error = %v", err)
	}
	if !reflect.DeepEqual(payload, want) {
		t.Errorf("BuildJSON() = %s, want %s", payload, want)
	}
}

func TestJSONEscaping(t *testing.T) {
	t.Parallel()

	want := Embed{
		Title:       "quotes: \" slash: \\ control: \u007f",
		Description: "html: <>& separator: \u2028 emoji: 🙂",
		Fields:      []Field{{Name: "line\nname", Value: "tab\tvalue"}},
	}
	payload, err := NewFromEmbed(want).BuildJSON()
	if err != nil {
		t.Fatalf("BuildJSON() error = %v", err)
	}

	var decoded Embed
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; payload = %s", err, payload)
	}
	if !reflect.DeepEqual(decoded, want) {
		t.Errorf("decoded JSON = %#v, want %#v", decoded, want)
	}
}
