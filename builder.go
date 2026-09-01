package embedbuilder

import "time"

const initialFieldCapacity = 4

type characterCounts struct {
	title           int
	description     int
	footer          int
	author          int
	fields          int
	total           int
	fieldViolations []Violation
	ruleViolations  []Violation
	rulesDirty      bool
}

// Builder incrementally constructs a Discord embed. A Builder is not safe for concurrent use.
type Builder struct {
	embed  Embed
	counts characterCounts
}

// New returns an empty Builder with capacity tuned for common small embeds.
func New() *Builder {
	return &Builder{
		embed: Embed{Fields: make([]Field, 0, initialFieldCapacity)},
	}
}

// SetTitle sets the embed title.
func (b *Builder) SetTitle(title string) *Builder {
	characters := characterCount(title)
	b.counts.total += characters - b.counts.title
	b.counts.title = characters
	b.embed.Title = title
	b.counts.rulesDirty = true

	return b
}

// SetDescription sets the embed description.
func (b *Builder) SetDescription(description string) *Builder {
	characters := characterCount(description)
	b.counts.total += characters - b.counts.description
	b.counts.description = characters
	b.embed.Description = description
	b.counts.rulesDirty = true

	return b
}

// SetURL sets the URL opened when the embed title is selected.
func (b *Builder) SetURL(url string) *Builder {
	b.embed.URL = url
	b.counts.rulesDirty = true

	return b
}

// SetTimestamp sets the embed timestamp in RFC 3339 format. Discord accepts ISO 8601 timestamps; UTC gives a stable
// representation across frameworks.
func (b *Builder) SetTimestamp(timestamp time.Time) *Builder {
	b.embed.Timestamp = timestamp.UTC().Format(time.RFC3339Nano)
	b.counts.rulesDirty = true

	return b
}

// WithCurrentTimestamp sets the embed timestamp to the current time.
func (b *Builder) WithCurrentTimestamp() *Builder {
	return b.SetTimestamp(time.Now())
}

// ClearTimestamp removes the embed timestamp.
func (b *Builder) ClearTimestamp() *Builder {
	b.embed.Timestamp = ""
	b.counts.rulesDirty = true

	return b
}

// SetColor sets the embed accent color as an RGB integer.
func (b *Builder) SetColor(color uint32) *Builder {
	if b.embed.Color != nil {
		*b.embed.Color = color

		return b
	}

	return b.setInitialColor(color)
}

// ClearColor removes the embed accent color.
func (b *Builder) ClearColor() *Builder {
	b.embed.Color = nil

	return b
}

// SetFooter sets the footer text and optional icon URL.
func (b *Builder) SetFooter(text string, iconURL ...string) *Builder {
	footer := b.embed.Footer
	if footer == nil {
		footer = b.setInitialFooter(text)
	} else {
		*footer = Footer{Text: text}
	}

	if len(iconURL) != 0 {
		footer.IconURL = iconURL[0]
	}

	characters := characterCount(text)
	b.counts.total += characters - b.counts.footer
	b.counts.footer = characters
	b.embed.Footer = footer
	b.counts.rulesDirty = true

	return b
}

// ClearFooter removes the embed footer.
func (b *Builder) ClearFooter() *Builder {
	b.counts.total -= b.counts.footer
	b.counts.footer = 0
	b.embed.Footer = nil
	b.counts.rulesDirty = true

	return b
}

// SetAuthor sets the author name and optional URL and icon URL. Optional arguments are interpreted in that order.
func (b *Builder) SetAuthor(name string, optional ...string) *Builder {
	author := b.embed.Author
	if author == nil {
		author = b.setInitialAuthor(name)
	} else {
		*author = Author{Name: name}
	}

	if len(optional) > 0 {
		author.URL = optional[0]
	}

	if len(optional) > 1 {
		author.IconURL = optional[1]
	}

	characters := characterCount(name)
	b.counts.total += characters - b.counts.author
	b.counts.author = characters
	b.embed.Author = author
	b.counts.rulesDirty = true

	return b
}

// ClearAuthor removes the embed author.
func (b *Builder) ClearAuthor() *Builder {
	b.counts.total -= b.counts.author
	b.counts.author = 0
	b.embed.Author = nil
	b.counts.rulesDirty = true

	return b
}

// SetImage sets the large embed image URL.
func (b *Builder) SetImage(url string) *Builder {
	if b.embed.Image == nil {
		b.setInitialImage(url)
	} else {
		b.embed.Image.URL = url
	}

	b.counts.rulesDirty = true

	return b
}

// ClearImage removes the large embed image.
func (b *Builder) ClearImage() *Builder {
	b.embed.Image = nil
	b.counts.rulesDirty = true

	return b
}

// SetThumbnail sets the embed thumbnail URL.
func (b *Builder) SetThumbnail(url string) *Builder {
	if b.embed.Thumbnail == nil {
		b.setInitialThumbnail(url)
	} else {
		b.embed.Thumbnail.URL = url
	}

	b.counts.rulesDirty = true

	return b
}

// ClearThumbnail removes the embed thumbnail.
func (b *Builder) ClearThumbnail() *Builder {
	b.embed.Thumbnail = nil
	b.counts.rulesDirty = true

	return b
}

// AddField appends a field.
func (b *Builder) AddField(name, value string, inline bool) *Builder {
	b.addField(Field{
		Name:   name,
		Value:  value,
		Inline: inline,
	})

	return b
}

// AddFields appends fields in order.
func (b *Builder) AddFields(fields ...Field) *Builder {
	for _, field := range fields {
		b.addField(field)
	}

	return b
}

// ClearFields removes all fields while retaining allocated storage for reuse.
func (b *Builder) ClearFields() *Builder {
	b.counts.total -= b.counts.fields
	b.counts.fields = 0
	b.counts.fieldViolations = b.counts.fieldViolations[:0]
	b.embed.Fields = b.embed.Fields[:0]
	b.counts.rulesDirty = true

	return b
}

// Validate checks all Discord writable constraints and limits before an embed
// is built.
func (b *Builder) Validate() error {
	if b == nil {
		return ErrNilBuilder
	}

	return validateBuilder(b)
}

// Build validates and returns an independent snapshot of the embed. Mutating the Builder after Build does not change
// the returned value.
func (b *Builder) Build() (Embed, error) {
	if err := b.Validate(); err != nil {
		return Embed{}, err
	}

	return cloneEmbed(b.embed), nil
}

func cloneEmbed(source Embed) Embed {
	result := source

	if source.Color != nil {
		color := *source.Color
		result.Color = &color
	}
	if source.Footer != nil {
		footer := *source.Footer
		result.Footer = &footer
	}
	if source.Image != nil {
		image := *source.Image
		result.Image = &image
	}
	if source.Thumbnail != nil {
		thumbnail := *source.Thumbnail
		result.Thumbnail = &thumbnail
	}
	if source.Author != nil {
		author := *source.Author
		result.Author = &author
	}
	if source.Fields != nil {
		result.Fields = append([]Field(nil), source.Fields...)
	}

	return result
}

func (b *Builder) setInitialColor(color uint32) *Builder {
	b.embed.Color = &color

	return b
}

func (b *Builder) setInitialFooter(text string) *Footer {
	b.embed.Footer = &Footer{Text: text}

	return b.embed.Footer
}

func (b *Builder) setInitialAuthor(name string) *Author {
	b.embed.Author = &Author{Name: name}

	return b.embed.Author
}

func (b *Builder) setInitialImage(url string) {
	b.embed.Image = &Media{URL: url}
}

func (b *Builder) setInitialThumbnail(url string) {
	b.embed.Thumbnail = &Media{URL: url}
}

func (b *Builder) addField(field Field) {
	index := len(b.embed.Fields)
	nameCharacters := characterCount(field.Name)
	valueCharacters := characterCount(field.Value)
	fieldCharacters := nameCharacters + valueCharacters

	b.ensureFieldCapacity()

	b.embed.Fields = append(b.embed.Fields, field)
	b.counts.fields += fieldCharacters
	b.counts.total += fieldCharacters
	b.counts.rulesDirty = true

	if nameCharacters > MaxFieldNameCharacters {
		b.counts.fieldViolations = append(b.counts.fieldViolations, Violation{
			Path:   fieldViolationPath(index, "name"),
			Actual: nameCharacters,
			Limit:  MaxFieldNameCharacters,
		})
	}
	if valueCharacters > MaxFieldValueCharacters {
		b.counts.fieldViolations = append(b.counts.fieldViolations, Violation{
			Path:   fieldViolationPath(index, "value"),
			Actual: valueCharacters,
			Limit:  MaxFieldValueCharacters,
		})
	}
}

func (b *Builder) ensureFieldCapacity() {
	if len(b.embed.Fields) < cap(b.embed.Fields) || cap(b.embed.Fields) >= MaxFields {
		return
	}

	capacity := initialFieldCapacity

	if cap(b.embed.Fields) != 0 {
		capacity = MaxFields
	}

	fields := make([]Field, len(b.embed.Fields), capacity)

	copy(fields, b.embed.Fields)

	b.embed.Fields = fields
}
