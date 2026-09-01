package embedbuilder

import (
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCharacterBudget(t *testing.T) {
	t.Parallel()

	builder := New().
		SetTitle(" title ").
		SetDescription("description").
		AddField("name", "value", false)

	wantCharacters := len("title") + len("description") + len("name") + len("value")
	if builder.CharacterCount() != wantCharacters {
		t.Errorf("CharacterCount() = %d, want %d", builder.CharacterCount(), wantCharacters)
	}
	if builder.RemainingCharacters() != MaxTotalCharacters-wantCharacters {
		t.Errorf("RemainingCharacters() = %d", builder.RemainingCharacters())
	}
	if builder.FieldCount() != 1 {
		t.Errorf("FieldCount() = %d, want 1", builder.FieldCount())
	}
	if !builder.CanAddField("next", "value") {
		t.Error("CanAddField() = false for a valid field")
	}
	if builder.CanAddField("", "value") || builder.CanAddField("name", " ") {
		t.Error("CanAddField() = true for an empty required value")
	}
	if builder.CanAddField(strings.Repeat("x", MaxFieldNameCharacters+1), "value") {
		t.Error("CanAddField() = true for an oversized name")
	}

	full := New()
	for index := 0; index < MaxFields; index++ {
		full.AddField("name", "value", false)
	}
	if full.CanAddField("name", "value") {
		t.Error("CanAddField() = true after reaching the field limit")
	}
}

func TestCharacterBudgetNilAndOverLimit(t *testing.T) {
	t.Parallel()

	var builder *Builder
	if builder.CharacterCount() != 0 || builder.RemainingCharacters() != MaxTotalCharacters || builder.FieldCount() != 0 {
		t.Fatal("nil Builder budget accessors returned unexpected values")
	}
	if builder.CanAddField("name", "value") {
		t.Error("nil Builder CanAddField() = true")
	}

	overLimit := New().SetDescription(strings.Repeat("x", MaxDescriptionCharacters+1))
	if overLimit.CanAddField("name", "value") {
		t.Error("invalid Builder CanAddField() = true")
	}
	if overLimit.RemainingCharacters() != MaxTotalCharacters-(MaxDescriptionCharacters+1) {
		t.Errorf("RemainingCharacters() = %d", overLimit.RemainingCharacters())
	}

	beyondTotal := New().
		SetDescription(strings.Repeat("d", MaxDescriptionCharacters)).
		SetFooter(strings.Repeat("f", MaxFooterCharacters)).
		SetTitle("x")
	if beyondTotal.RemainingCharacters() != 0 {
		t.Errorf("RemainingCharacters() = %d, want 0", beyondTotal.RemainingCharacters())
	}
}

func TestFieldMutation(t *testing.T) {
	t.Parallel()

	builder := New().
		AddField("zero", "0", false).
		AddField("two", "2", false)

	if err := builder.InsertField(1, "one", "1", true); err != nil {
		t.Fatalf("InsertField() error = %v", err)
	}
	if err := builder.SetField(2, "second", "updated", true); err != nil {
		t.Fatalf("SetField() error = %v", err)
	}
	if err := builder.RemoveField(0); err != nil {
		t.Fatalf("RemoveField() error = %v", err)
	}

	embed, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	want := []Field{
		{Name: "one", Value: "1", Inline: true},
		{Name: "second", Value: "updated", Inline: true},
	}
	if !reflect.DeepEqual(embed.Fields, want) {
		t.Errorf("Build() fields = %#v, want %#v", embed.Fields, want)
	}
	if builder.CharacterCount() != len("one1secondupdated") {
		t.Errorf("CharacterCount() = %d", builder.CharacterCount())
	}
}

func TestFieldMutationIndexErrors(t *testing.T) {
	t.Parallel()

	builder := New().AddField("name", "value", false)
	tests := []error{
		builder.SetField(-1, "name", "value", false),
		builder.SetField(1, "name", "value", false),
		builder.InsertField(2, "name", "value", false),
		builder.RemoveField(1),
	}

	for _, err := range tests {
		if !errors.Is(err, ErrFieldIndexOutOfRange) {
			t.Errorf("mutation error = %v, want ErrFieldIndexOutOfRange", err)
		}
	}

	var nilBuilder *Builder
	if !errors.Is(nilBuilder.RemoveField(0), ErrNilBuilder) {
		t.Errorf("nil RemoveField() error = %v", nilBuilder.RemoveField(0))
	}
}

func TestNewFromEmbedCloneAndReset(t *testing.T) {
	t.Parallel()

	color := ColorBlurple
	source := Embed{
		Title:  "source",
		Color:  &color,
		Footer: &Footer{Text: "footer"},
		Fields: []Field{{Name: "name", Value: "value"}},
	}
	builder := NewFromEmbed(source)
	clone := builder.Clone()

	source.Title = "changed"
	*source.Color = 0
	source.Footer.Text = "changed"
	source.Fields[0].Name = "changed"

	first, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if first.Title != "source" || *first.Color != ColorBlurple || first.Footer.Text != "footer" || first.Fields[0].Name != "name" {
		t.Fatalf("NewFromEmbed() retained source aliases: %#v", first)
	}

	clone.SetTitle("clone")
	second, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if second.Title != "source" {
		t.Errorf("Clone() mutation changed source Builder: %#v", second)
	}

	builder.Reset()
	reset, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() after Reset error = %v", err)
	}
	if !reflect.DeepEqual(reset, Embed{Fields: []Field(nil)}) && !reflect.DeepEqual(reset, Embed{Fields: []Field{}}) {
		t.Errorf("Reset() Build() = %#v", reset)
	}
	if builder.CharacterCount() != 0 || builder.FieldCount() != 0 {
		t.Errorf("Reset() counts = %d, %d", builder.CharacterCount(), builder.FieldCount())
	}

	var nilBuilder *Builder
	if nilBuilder.Clone() != nil || nilBuilder.Reset() != nil {
		t.Error("nil lifecycle method did not return nil")
	}
}

func TestColorHelpers(t *testing.T) {
	t.Parallel()

	builder := New().SetColorRGB(0x12, 0x34, 0x56)
	rgbEmbed, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() RGB error = %v", err)
	}
	if rgbEmbed.Color == nil || *rgbEmbed.Color != 0x123456 {
		t.Errorf("Build() RGB color = %v", rgbEmbed.Color)
	}

	if err := builder.SetColorHex("#aBcDeF"); err != nil {
		t.Fatalf("SetColorHex() error = %v", err)
	}

	embed, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}
	if embed.Color == nil || *embed.Color != 0xABCDEF {
		t.Errorf("Build() color = %v", embed.Color)
	}

	if err := builder.SetColorHex("123456"); err != nil {
		t.Fatalf("SetColorHex() without hash error = %v", err)
	}
	if err := builder.SetColorHex("#xyzxyz"); !errors.Is(err, ErrInvalidHexColor) {
		t.Errorf("SetColorHex() error = %v, want ErrInvalidHexColor", err)
	}
	unchanged, err := builder.Build()
	if err != nil {
		t.Fatalf("Build() after invalid hex error = %v", err)
	}
	if unchanged.Color == nil || *unchanged.Color != 0x123456 {
		t.Errorf("invalid SetColorHex() changed color to %v", unchanged.Color)
	}
}

func TestSetColorHexShortForm(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input string
		want  uint32
	}{
		{input: "321", want: 0x332211},
		{input: "#aBc", want: 0xAABBCC},
	}

	for _, test := range tests {
		test := test
		t.Run(test.input, func(t *testing.T) {
			t.Parallel()

			builder := New()
			if err := builder.SetColorHex(test.input); err != nil {
				t.Fatalf("SetColorHex(%q) error = %v", test.input, err)
			}

			embed, err := builder.Build()
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}
			if embed.Color == nil || *embed.Color != test.want {
				t.Errorf("SetColorHex(%q) color = %v, want %#06x", test.input, embed.Color, test.want)
			}
		})
	}
}
