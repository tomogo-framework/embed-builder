package embedbuilder

import (
	"strings"
	"testing"
)

var (
	benchmarkEmbed  Embed
	benchmarkEmbeds []Embed
	benchmarkBytes  []byte
	benchmarkError  error
)

func BenchmarkBuild(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		benchmarkBuild(b, newSmallBenchmarkBuilder())
	})
	b.Run("Maximum", func(b *testing.B) {
		benchmarkBuild(b, newMaximumBenchmarkBuilder())
	})
}

func BenchmarkValidate(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		benchmarkValidate(b, newSmallBenchmarkBuilder())
	})
	b.Run("MaximumASCII", func(b *testing.B) {
		benchmarkValidate(b, newMaximumBenchmarkBuilder())
	})
	b.Run("Unicode", func(b *testing.B) {
		builder := New().SetDescription(strings.Repeat("🙂", MaxDescriptionCharacters))

		benchmarkValidate(b, builder)
	})
}

func BenchmarkValidateEmbeds(b *testing.B) {
	embeds := make([]Embed, MaxEmbedsPerMessage)
	for index := range embeds {
		embeds[index].Description = strings.Repeat("x", 500)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		benchmarkError = ValidateEmbeds(embeds...)
	}
}

func BenchmarkComposeAndBuild(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		b.ReportAllocs()

		for b.Loop() {
			builder := newSmallBenchmarkBuilder()

			benchmarkEmbed, benchmarkError = builder.Build()
		}
	})
	b.Run("Maximum", func(b *testing.B) {
		title := strings.Repeat("t", 100)
		description := strings.Repeat("d", 3000)
		footer := strings.Repeat("f", 25)
		author := strings.Repeat("a", 25)
		name := strings.Repeat("n", 10)
		value := strings.Repeat("v", 100)

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			builder := New().
				SetTitle(title).
				SetDescription(description).
				SetColor(0x5865F2).
				SetFooter(footer).
				SetAuthor(author)
			for index := 0; index < MaxFields; index++ {
				builder.AddField(name, value, index%3 != 0)
			}

			benchmarkEmbed, benchmarkError = builder.Build()
		}
	})
}

func BenchmarkBuildJSON(b *testing.B) {
	b.Run("Small", func(b *testing.B) {
		benchmarkJSON(b, newSmallBenchmarkBuilder(), false)
	})
	b.Run("Maximum", func(b *testing.B) {
		benchmarkJSON(b, newMaximumBenchmarkBuilder(), false)
	})
	b.Run("AppendMaximum", func(b *testing.B) {
		benchmarkJSON(b, newMaximumBenchmarkBuilder(), true)
	})
}

func BenchmarkFieldMutation(b *testing.B) {
	b.Run("SetMaximum", func(b *testing.B) {
		builder := newMaximumBenchmarkBuilder()
		index := 0

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			benchmarkError = builder.SetField(index, "updated", "updated value", true)
			index++
			if index == MaxFields {
				index = 0
			}
		}
	})
	b.Run("InsertRemoveMaximum", func(b *testing.B) {
		builder := newMaximumBenchmarkBuilder()

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			benchmarkError = builder.InsertField(12, "inserted", "inserted value", false)
			if benchmarkError != nil {
				b.Fatal(benchmarkError)
			}
			benchmarkError = builder.RemoveField(12)
		}
	})
}

func BenchmarkSetColorHex(b *testing.B) {
	b.Run("Short", func(b *testing.B) {
		builder := New().SetColor(0)

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			benchmarkError = builder.SetColorHex("321")
		}
	})
	b.Run("Long", func(b *testing.B) {
		builder := New().SetColor(0)

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			benchmarkError = builder.SetColorHex("332211")
		}
	})
}

func BenchmarkSetOptionalMetadata(b *testing.B) {
	b.Run("Footer", func(b *testing.B) {
		builder := New().SetFooter("initial")

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			builder.SetFooter("footer", "https://example.com/footer.png")
		}
	})
	b.Run("Author", func(b *testing.B) {
		builder := New().SetAuthor("initial")

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			builder.SetAuthor("author", "https://example.com", "https://example.com/author.png")
		}
	})
	b.Run("Image", func(b *testing.B) {
		builder := New().SetImage("https://example.com/initial.png")

		b.ReportAllocs()
		b.ResetTimer()

		for b.Loop() {
			builder.SetImage("https://example.com/image.png")
		}
	})
}

func BenchmarkCollectionBuild(b *testing.B) {
	collection := NewCollection()
	for index := 0; index < MaxEmbedsPerMessage; index++ {
		collection.Add(Embed{Description: strings.Repeat("x", 500)})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		benchmarkEmbeds, benchmarkError = collection.Build()
	}
}

func BenchmarkCollectionBuildJSON(b *testing.B) {
	collection := NewCollection()
	for index := 0; index < MaxEmbedsPerMessage; index++ {
		collection.Add(Embed{Description: strings.Repeat("x", 500)})
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		benchmarkBytes, benchmarkError = collection.BuildJSON()
	}
}

func BenchmarkCloneMaximum(b *testing.B) {
	builder := newMaximumBenchmarkBuilder()

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		_ = builder.Clone()
	}
}

func benchmarkBuild(b *testing.B, builder *Builder) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		var err error
		benchmarkEmbed, err = builder.Build()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func benchmarkValidate(b *testing.B, builder *Builder) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		benchmarkError = builder.Validate()
	}
}

func benchmarkJSON(b *testing.B, builder *Builder, appendMode bool) {
	b.Helper()
	b.ReportAllocs()
	b.ResetTimer()

	if appendMode {
		buffer := make([]byte, 0, 16*1024)
		for b.Loop() {
			benchmarkBytes, benchmarkError = builder.AppendJSON(buffer[:0])
		}

		return
	}

	for b.Loop() {
		benchmarkBytes, benchmarkError = builder.BuildJSON()
	}
}

func newSmallBenchmarkBuilder() *Builder {
	return New().
		SetTitle("Deployment complete").
		SetDescription("Version 2.4.0 is live and healthy.").
		SetColor(0x57F287).
		SetFooter("Automated notification").
		SetAuthor("Release bot").
		AddField("Environment", "production", true).
		AddField("Status", "healthy", true)
}

func newMaximumBenchmarkBuilder() *Builder {
	builder := New().
		SetTitle(strings.Repeat("t", 100)).
		SetDescription(strings.Repeat("d", 3000)).
		SetColor(0x5865F2).
		SetFooter(strings.Repeat("f", 25)).
		SetAuthor(strings.Repeat("a", 25))
	for index := 0; index < MaxFields; index++ {
		builder.AddField(strings.Repeat("n", 10), strings.Repeat("v", 100), index%3 != 0)
	}

	return builder
}
