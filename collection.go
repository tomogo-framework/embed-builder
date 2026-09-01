package embedbuilder

// CollectionBuilder constructs the complete embed array for one Discord message and validates limits shared across that
// array.
type CollectionBuilder struct {
	embeds    []Embed
	validated bool
}

// NewCollection returns an empty message-level embed collection.
func NewCollection() *CollectionBuilder {
	return &CollectionBuilder{
		embeds:    make([]Embed, 0, 2),
		validated: true,
	}
}

// Add appends an independent copy of embed.
func (c *CollectionBuilder) Add(embed Embed) *CollectionBuilder {
	c.embeds = append(c.embeds, cloneEmbed(embed))
	c.validated = false

	return c
}

// AddBuilder validates and appends a snapshot from builder.
func (c *CollectionBuilder) AddBuilder(builder *Builder) error {
	if c == nil {
		return ErrNilCollection
	}

	embed, err := builder.Build()

	if err != nil {
		return err
	}

	c.embeds = append(c.embeds, embed)
	c.validated = false

	return nil
}

// Count returns the number of embeds in the collection.
func (c *CollectionBuilder) Count() int {
	if c == nil {
		return 0
	}

	return len(c.embeds)
}

// Clear removes all embeds while retaining allocated storage for reuse.
func (c *CollectionBuilder) Clear() *CollectionBuilder {
	if c == nil {
		return nil
	}

	clear(c.embeds)

	c.embeds = c.embeds[:0]
	c.validated = true

	return c
}

// Validate checks all per-embed and shared message limits.
func (c *CollectionBuilder) Validate() error {
	if c == nil {
		return ErrNilCollection
	}
	if c.validated {
		return nil
	}

	err := ValidateEmbeds(c.embeds...)
	if err == nil {
		c.validated = true
	}

	return err
}

// Build validates and returns independent snapshots of all embeds.
func (c *CollectionBuilder) Build() ([]Embed, error) {
	if err := c.Validate(); err != nil {
		return nil, err
	}

	embeds := make([]Embed, len(c.embeds))
	for index, embed := range c.embeds {
		embeds[index] = cloneEmbed(embed)
	}

	return embeds, nil
}
