package embedbuilder

import "testing"

var benchmarkBuilder *Builder

func BenchmarkClone(b *testing.B) {
	for _, test := range []struct {
		name    string
		builder *Builder
	}{
		{
			name:    "Small",
			builder: newSmallBenchmarkBuilder(),
		},
		{
			name:    "MaximumDirty",
			builder: newMaximumBenchmarkBuilder(),
		},
		{
			name:    "MaximumValidated",
			builder: newMaximumBenchmarkBuilder(),
		},
	} {
		b.Run(test.name, func(b *testing.B) {
			if test.name == "MaximumValidated" {
				if err := test.builder.Validate(); err != nil {
					b.Fatal(err)
				}
			}

			b.ReportAllocs()
			for b.Loop() {
				benchmarkBuilder = test.builder.Clone()
			}
		})
	}
}

func BenchmarkNewFromEmbedSmall(b *testing.B) {
	embed, err := newSmallBenchmarkBuilder().Build()
	if err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	for b.Loop() {
		benchmarkBuilder = NewFromEmbed(embed)
	}
}
