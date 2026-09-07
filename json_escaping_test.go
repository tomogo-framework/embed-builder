package embedbuilder

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestJSONEscapingAcrossProperties(t *testing.T) {
	t.Parallel()

	var ascii strings.Builder
	for character := 0; character < 128; character++ {
		ascii.WriteByte(byte(character))
	}

	value := "x" + ascii.String() + "🙂\u2028\u2029x"
	url := "https://example.com/\"quoted\""
	color := uint32(MaxColor)
	embed := Embed{
		Title:       value,
		Description: value,
		URL:         url,
		Timestamp:   "2026-09-07T00:00:00Z",
		Color:       &color,
		Footer: &Footer{
			Text:    value,
			IconURL: url,
		},
		Author: &Author{
			Name:    value,
			URL:     url,
			IconURL: url,
		},
		Image: &Media{
			URL: url,
		},
		Thumbnail: &Media{
			URL: url,
		},
		Fields: []Field{
			{
				Name:   value,
				Value:  value,
				Inline: true,
			},
			{
				Name:  "plain",
				Value: "plain",
			},
		},
	}
	builder := NewFromEmbed(embed)
	payload, err := builder.BuildJSON()
	if err != nil {
		t.Fatal(err)
	}

	var decoded Embed
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(decoded, embed) {
		t.Fatalf("JSON changed embed: %#v", decoded)
	}

	prefix := []byte("prefix:")
	buffer := make([]byte, len(prefix), len(prefix)+len(payload))
	copy(buffer, prefix)
	appended, err := builder.AppendJSON(buffer)
	if err != nil {
		t.Fatal(err)
	}

	if &appended[0] != &buffer[0] || string(appended) != string(prefix)+string(payload) {
		t.Fatal("AppendJSON failed to reuse an exactly sized buffer or preserve its prefix")
	}
}

func TestEscapedJSONAllocations(t *testing.T) {
	value := strings.Repeat("\x00", 500)
	builder := New().SetDescription(value)
	collection := NewCollection()
	for index := 0; index < MaxEmbedsPerMessage; index++ {
		collection.Add(Embed{
			Description: value,
		})
	}

	for _, build := range []func() ([]byte, error){builder.BuildJSON, collection.BuildJSON} {
		if _, err := build(); err != nil {
			t.Fatal(err)
		}

		allocations := testing.AllocsPerRun(100, func() {
			benchmarkBytes, benchmarkError = build()
		})
		if benchmarkError != nil {
			t.Fatal(benchmarkError)
		}

		// Race instrumentation can add an allocation beyond the output buffer.
		if allocations > 2 {
			t.Errorf("escaped JSON allocated %g times, want at most two including race instrumentation", allocations)
		}
	}
}

func FuzzBuildJSONRoundTrip(f *testing.F) {
	for _, value := range []string{"plain", "\x00\b\f\n\r\t\"\\", "<>&🙂\u2028\u2029", "\xff"} {
		f.Add(value)
	}

	f.Fuzz(func(t *testing.T, value string) {
		builder := New().SetDescription(value)
		payload, err := builder.BuildJSON()
		if err != nil {
			return
		}

		var decoded Embed
		if err := json.Unmarshal(payload, &decoded); err != nil {
			t.Fatal(err)
		}

		if decoded.Description != value {
			t.Fatalf("JSON changed description: %q != %q", decoded.Description, value)
		}

		plain := Embed{
			Title: "plain",
		}
		collection := NewCollection().Add(plain).Add(decoded).Add(plain)
		payload, err = collection.BuildJSON()
		if err != nil {
			t.Fatal(err)
		}

		var embeds []Embed
		if err := json.Unmarshal(payload, &embeds); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(embeds, []Embed{plain, decoded, plain}) {
			t.Fatalf("collection JSON changed embeds: %#v", embeds)
		}
	})
}
