package embedbuilder

import "slices"

func newBuilderSnapshot(embed Embed) *Builder {
	fields := embed.Fields
	embed.Fields = nil
	builder := &Builder{
		embed: cloneEmbed(embed),
	}
	builder.embed.Fields = make([]Field, len(fields), max(initialFieldCapacity, len(fields)))
	copy(builder.embed.Fields, fields)

	return builder
}

// NewFromEmbed returns a Builder initialized from an independent copy of embed. Validation remains deferred until
// Validate or Build.
func NewFromEmbed(embed Embed) *Builder {
	builder := newBuilderSnapshot(embed)
	builder.recalculateAllCounts()
	builder.counts.rulesDirty = true

	return builder
}

// Clone returns an independent copy of the Builder. A nil Builder clones to nil.
func (b *Builder) Clone() *Builder {
	if b == nil {
		return nil
	}

	clone := newBuilderSnapshot(b.embed)
	clone.counts = b.counts
	clone.counts.fieldViolations = slices.Clone(b.counts.fieldViolations)
	clone.counts.ruleViolations = slices.Clone(b.counts.ruleViolations)

	return clone
}

// Reset clears all embed content while retaining field storage for reuse.
func (b *Builder) Reset() *Builder {
	if b == nil {
		return nil
	}

	fields := b.embed.Fields
	clear(fields)
	b.embed = Embed{Fields: fields[:0]}
	b.counts = characterCounts{}

	return b
}

func (b *Builder) recalculateAllCounts() {
	b.counts = characterCounts{
		title:       characterCount(b.embed.Title),
		description: characterCount(b.embed.Description),
	}
	b.counts.total = b.counts.title + b.counts.description
	if b.embed.Footer != nil {
		b.counts.footer = characterCount(b.embed.Footer.Text)
		b.counts.total += b.counts.footer
	}
	if b.embed.Author != nil {
		b.counts.author = characterCount(b.embed.Author.Name)
		b.counts.total += b.counts.author
	}

	b.recalculateFields()
}
