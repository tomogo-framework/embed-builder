package embedbuilder

// NewFromEmbed returns a Builder initialized from an independent copy of embed. Validation remains deferred until
// Validate or Build.
func NewFromEmbed(embed Embed) *Builder {
	builder := &Builder{embed: cloneEmbed(embed)}
	if cap(builder.embed.Fields) < initialFieldCapacity {
		fields := make([]Field, len(builder.embed.Fields), initialFieldCapacity)
		copy(fields, builder.embed.Fields)
		builder.embed.Fields = fields
	}

	builder.recalculateAllCounts()
	builder.counts.rulesDirty = true

	return builder
}

// Clone returns an independent copy of the Builder. A nil Builder clones to nil.
func (b *Builder) Clone() *Builder {
	if b == nil {
		return nil
	}

	return NewFromEmbed(b.embed)
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
