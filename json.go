package embedbuilder

import (
	"encoding/json/jsontext"
	"slices"
	"strconv"
)

// BuildJSON validates and marshals the embed as a Discord-compatible JSON object.
func (b *Builder) BuildJSON() ([]byte, error) {
	return b.AppendJSON(nil)
}

// AppendJSON validates and appends a Discord-compatible JSON object to dst.
func (b *Builder) AppendJSON(dst []byte) ([]byte, error) {
	if err := b.Validate(); err != nil {
		return dst, err
	}

	dst = slices.Grow(dst, estimateEmbedJSONSize(b.embed))

	return appendEmbedJSON(dst, b.embed), nil
}

// BuildJSON validates and marshals the collection as a JSON embed array.
func (c *CollectionBuilder) BuildJSON() ([]byte, error) {
	return c.AppendJSON(nil)
}

// AppendJSON validates and appends the collection's JSON embed array to dst.
func (c *CollectionBuilder) AppendJSON(dst []byte) ([]byte, error) {
	if err := c.Validate(); err != nil {
		return dst, err
	}

	size := 2
	for _, embed := range c.embeds {
		size += estimateEmbedJSONSize(embed) + 1
	}

	dst = slices.Grow(dst, size)
	dst = append(dst, '[')
	for index, embed := range c.embeds {
		if index != 0 {
			dst = append(dst, ',')
		}

		dst = appendEmbedJSON(dst, embed)
	}

	dst = append(dst, ']')

	return dst, nil
}

func appendEmbedJSON(dst []byte, embed Embed) []byte {
	dst = append(dst, '{')

	first := true

	if embed.Title != "" {
		dst, first = appendStringProperty(dst, first, "title", embed.Title)
	}
	if embed.Description != "" {
		dst, first = appendStringProperty(dst, first, "description", embed.Description)
	}
	if embed.URL != "" {
		dst, first = appendStringProperty(dst, first, "url", embed.URL)
	}
	if embed.Timestamp != "" {
		dst, first = appendStringProperty(dst, first, "timestamp", embed.Timestamp)
	}
	if embed.Color != nil {
		dst, first = appendPropertyName(dst, first, "color")
		dst = strconv.AppendUint(dst, uint64(*embed.Color), 10)
	}
	if embed.Footer != nil {
		dst, first = appendPropertyName(dst, first, "footer")
		dst = appendFooterJSON(dst, *embed.Footer)
	}
	if embed.Image != nil {
		dst, first = appendPropertyName(dst, first, "image")
		dst = appendMediaJSON(dst, *embed.Image)
	}
	if embed.Thumbnail != nil {
		dst, first = appendPropertyName(dst, first, "thumbnail")
		dst = appendMediaJSON(dst, *embed.Thumbnail)
	}
	if embed.Author != nil {
		dst, first = appendPropertyName(dst, first, "author")
		dst = appendAuthorJSON(dst, *embed.Author)
	}
	if len(embed.Fields) != 0 {
		dst, first = appendPropertyName(dst, first, "fields")
		dst = appendFieldsJSON(dst, embed.Fields)
	}

	dst = append(dst, '}')

	return dst
}

func appendFooterJSON(dst []byte, footer Footer) []byte {
	dst = append(dst, '{')
	dst, _ = appendStringProperty(dst, true, "text", footer.Text)
	if footer.IconURL != "" {
		dst, _ = appendStringProperty(dst, false, "icon_url", footer.IconURL)
	}

	dst = append(dst, '}')

	return dst
}

func appendMediaJSON(dst []byte, media Media) []byte {
	dst = append(dst, '{')
	dst, _ = appendStringProperty(dst, true, "url", media.URL)
	dst = append(dst, '}')

	return dst
}

func appendAuthorJSON(dst []byte, author Author) []byte {
	dst = append(dst, '{')
	dst, _ = appendStringProperty(dst, true, "name", author.Name)

	if author.URL != "" {
		dst, _ = appendStringProperty(dst, false, "url", author.URL)
	}

	if author.IconURL != "" {
		dst, _ = appendStringProperty(dst, false, "icon_url", author.IconURL)
	}

	dst = append(dst, '}')

	return dst
}

func appendFieldsJSON(dst []byte, fields []Field) []byte {
	dst = append(dst, '[')
	for index, field := range fields {
		if index != 0 {
			dst = append(dst, ',')
		}

		dst = append(dst, '{')
		dst, _ = appendStringProperty(dst, true, "name", field.Name)
		dst, _ = appendStringProperty(dst, false, "value", field.Value)

		if field.Inline {
			dst, _ = appendPropertyName(dst, false, "inline")
			dst = append(dst, "true"...)
		}

		dst = append(dst, '}')
	}

	dst = append(dst, ']')

	return dst
}

func appendStringProperty(dst []byte, first bool, name, value string) ([]byte, bool) {
	dst, _ = appendPropertyName(dst, first, name)
	dst = appendJSONString(dst, value)

	return dst, false
}

func appendJSONString(dst []byte, value string) []byte {
	for index := 0; index < len(value); index++ {
		character := value[index]
		if character < ' ' || character == '"' || character == '\\' {
			dst, _ = jsontext.AppendQuote(dst, value)

			return dst
		}
	}

	dst = append(dst, '"')
	dst = append(dst, value...)
	dst = append(dst, '"')

	return dst
}

func appendPropertyName(dst []byte, first bool, name string) ([]byte, bool) {
	if !first {
		dst = append(dst, ',')
	}

	dst = append(dst, '"')
	dst = append(dst, name...)
	dst = append(dst, '"', ':')

	return dst, false
}

func estimateEmbedJSONSize(embed Embed) int {
	size := 2
	properties := 0

	if embed.Title != "" {
		size += len("title") + 5 + len(embed.Title)
		properties++
	}
	if embed.Description != "" {
		size += len("description") + 5 + len(embed.Description)
		properties++
	}
	if embed.URL != "" {
		size += len("url") + 5 + len(embed.URL)
		properties++
	}
	if embed.Timestamp != "" {
		size += len("timestamp") + 5 + len(embed.Timestamp)
		properties++
	}
	if embed.Color != nil {
		size += len("color") + 3 + 8
		properties++
	}

	if embed.Footer != nil {
		footerSize := 2 + len("text") + 5 + len(embed.Footer.Text)
		if embed.Footer.IconURL != "" {
			footerSize += 1 + len("icon_url") + 5 + len(embed.Footer.IconURL)
		}

		size += len("footer") + 3 + footerSize
		properties++
	}
	if embed.Image != nil {
		mediaSize := 2 + len("url") + 5 + len(embed.Image.URL)
		size += len("image") + 3 + mediaSize
		properties++
	}
	if embed.Thumbnail != nil {
		mediaSize := 2 + len("url") + 5 + len(embed.Thumbnail.URL)
		size += len("thumbnail") + 3 + mediaSize
		properties++
	}
	if embed.Author != nil {
		authorSize := 2 + len("name") + 5 + len(embed.Author.Name)
		if embed.Author.URL != "" {
			authorSize += 1 + len("url") + 5 + len(embed.Author.URL)
		}
		if embed.Author.IconURL != "" {
			authorSize += 1 + len("icon_url") + 5 + len(embed.Author.IconURL)
		}

		size += len("author") + 3 + authorSize
		properties++
	}

	if len(embed.Fields) != 0 {
		fieldsSize := 2 + len(embed.Fields) - 1
		for _, field := range embed.Fields {
			fieldSize := 2 + len("name") + 5 + len(field.Name)
			fieldSize += 1 + len("value") + 5 + len(field.Value)
			if field.Inline {
				fieldSize += 1 + len("inline") + 3 + len("true")
			}

			fieldsSize += fieldSize
		}

		size += len("fields") + 3 + fieldsSize
		properties++
	}

	if properties > 1 {
		size += properties - 1
	}

	return size
}
