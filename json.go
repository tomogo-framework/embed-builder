package embedbuilder

import (
	"slices"
	"strconv"
)

// Validated UTF-8 only expands when an ASCII byte needs a JSON escape.
var jsonEscapeExtraBytes = [256]uint8{
	0:    5,
	1:    5,
	2:    5,
	3:    5,
	4:    5,
	5:    5,
	6:    5,
	7:    5,
	8:    1,
	9:    1,
	10:   1,
	11:   5,
	12:   1,
	13:   1,
	14:   5,
	15:   5,
	16:   5,
	17:   5,
	18:   5,
	19:   5,
	20:   5,
	21:   5,
	22:   5,
	23:   5,
	24:   5,
	25:   5,
	26:   5,
	27:   5,
	28:   5,
	29:   5,
	30:   5,
	31:   5,
	'"':  1,
	'\\': 1,
}

// BuildJSON validates and marshals the embed as a Discord-compatible JSON object.
func (b *Builder) BuildJSON() ([]byte, error) {
	return b.AppendJSON(nil)
}

// AppendJSON validates and appends a Discord-compatible JSON object to dst.
func (b *Builder) AppendJSON(dst []byte) ([]byte, error) {
	if err := b.Validate(); err != nil {
		return dst, err
	}

	size, escaped := estimateEmbedJSONSize(b.embed)
	dst = slices.Grow(dst, size)

	return appendEmbedJSON(dst, b.embed, escaped), nil
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
	var escaped [MaxEmbedsPerMessage]bool
	for index, embed := range c.embeds {
		embedSize, needsEscaping := estimateEmbedJSONSize(embed)
		size += embedSize + 1
		escaped[index] = needsEscaping
	}

	dst = slices.Grow(dst, size)
	dst = append(dst, '[')
	for index, embed := range c.embeds {
		if index != 0 {
			dst = append(dst, ',')
		}

		dst = appendEmbedJSON(dst, embed, escaped[index])
	}

	dst = append(dst, ']')

	return dst, nil
}

func appendEmbedJSON(dst []byte, embed Embed, escaped bool) []byte {
	dst = append(dst, '{')

	first := true

	if embed.Title != "" {
		dst, first = appendStringProperty(dst, first, "title", embed.Title, escaped)
	}
	if embed.Description != "" {
		dst, first = appendStringProperty(dst, first, "description", embed.Description, escaped)
	}
	if embed.URL != "" {
		dst, first = appendStringProperty(dst, first, "url", embed.URL, escaped)
	}
	if embed.Timestamp != "" {
		dst, first = appendStringProperty(dst, first, "timestamp", embed.Timestamp, escaped)
	}
	if embed.Color != nil {
		dst, first = appendPropertyName(dst, first, "color")
		dst = strconv.AppendUint(dst, uint64(*embed.Color), 10)
	}
	if embed.Footer != nil {
		dst, first = appendPropertyName(dst, first, "footer")
		dst = appendFooterJSON(dst, *embed.Footer, escaped)
	}
	if embed.Image != nil {
		dst, first = appendPropertyName(dst, first, "image")
		dst = appendMediaJSON(dst, *embed.Image, escaped)
	}
	if embed.Thumbnail != nil {
		dst, first = appendPropertyName(dst, first, "thumbnail")
		dst = appendMediaJSON(dst, *embed.Thumbnail, escaped)
	}
	if embed.Author != nil {
		dst, first = appendPropertyName(dst, first, "author")
		dst = appendAuthorJSON(dst, *embed.Author, escaped)
	}
	if len(embed.Fields) != 0 {
		dst, first = appendPropertyName(dst, first, "fields")
		dst = appendFieldsJSON(dst, embed.Fields, escaped)
	}

	dst = append(dst, '}')

	return dst
}

func appendFooterJSON(dst []byte, footer Footer, escaped bool) []byte {
	dst = append(dst, '{')
	dst, _ = appendStringProperty(dst, true, "text", footer.Text, escaped)
	if footer.IconURL != "" {
		dst, _ = appendStringProperty(dst, false, "icon_url", footer.IconURL, escaped)
	}

	dst = append(dst, '}')

	return dst
}

func appendMediaJSON(dst []byte, media Media, escaped bool) []byte {
	dst = append(dst, '{')
	dst, _ = appendStringProperty(dst, true, "url", media.URL, escaped)
	dst = append(dst, '}')

	return dst
}

func appendAuthorJSON(dst []byte, author Author, escaped bool) []byte {
	dst = append(dst, '{')
	dst, _ = appendStringProperty(dst, true, "name", author.Name, escaped)

	if author.URL != "" {
		dst, _ = appendStringProperty(dst, false, "url", author.URL, escaped)
	}

	if author.IconURL != "" {
		dst, _ = appendStringProperty(dst, false, "icon_url", author.IconURL, escaped)
	}

	dst = append(dst, '}')

	return dst
}

func appendFieldsJSON(dst []byte, fields []Field, escaped bool) []byte {
	dst = append(dst, '[')
	for index, field := range fields {
		if index != 0 {
			dst = append(dst, ',')
		}

		dst = append(dst, '{')
		dst, _ = appendStringProperty(dst, true, "name", field.Name, escaped)
		dst, _ = appendStringProperty(dst, false, "value", field.Value, escaped)

		if field.Inline {
			dst, _ = appendPropertyName(dst, false, "inline")
			dst = append(dst, "true"...)
		}

		dst = append(dst, '}')
	}

	dst = append(dst, ']')

	return dst
}

func appendStringProperty(dst []byte, first bool, name, value string, escaped bool) ([]byte, bool) {
	dst, _ = appendPropertyName(dst, first, name)
	if escaped {
		dst = appendEscapedJSONString(dst, value)
	} else {
		dst = append(dst, '"')
		dst = append(dst, value...)
		dst = append(dst, '"')
	}

	return dst, false
}

// appendEscapedJSONString quotes UTF-8 that has already passed validation.
func appendEscapedJSONString(dst []byte, value string) []byte {
	const hex = "0123456789abcdef"
	dst = append(dst, '"')
	start := 0
	for index := 0; index < len(value); index++ {
		character := value[index]
		if jsonEscapeExtraBytes[character] == 0 {
			continue
		}

		if start < index {
			dst = append(dst, value[start:index]...)
		}

		switch character {
		case '"', '\\':
			dst = append(dst, '\\', character)
		case '\b':
			dst = append(dst, '\\', 'b')
		case '\f':
			dst = append(dst, '\\', 'f')
		case '\n':
			dst = append(dst, '\\', 'n')
		case '\r':
			dst = append(dst, '\\', 'r')
		case '\t':
			dst = append(dst, '\\', 't')
		default:
			dst = append(dst, '\\', 'u', '0', '0', hex[character>>4], hex[character&0xF])
		}

		start = index + 1
	}

	dst = append(dst, value[start:]...)
	dst = append(dst, '"')

	return dst
}

// jsonStringSize measures validated UTF-8 without the surrounding quotes.
func jsonStringSize(value string, escaped *bool) int {
	extra := 0
	index := 0
	for len(value)-index >= 4 {
		extra += int(jsonEscapeExtraBytes[value[index]]) + int(jsonEscapeExtraBytes[value[index+1]]) +
			int(jsonEscapeExtraBytes[value[index+2]]) + int(jsonEscapeExtraBytes[value[index+3]])
		index += 4
	}

	for ; index < len(value); index++ {
		extra += int(jsonEscapeExtraBytes[value[index]])
	}

	if extra != 0 {
		*escaped = true
	}

	return len(value) + extra
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

func estimateEmbedJSONSize(embed Embed) (int, bool) {
	var escaped bool
	size := 2
	properties := 0

	if embed.Title != "" {
		size += len("title") + 5 + jsonStringSize(embed.Title, &escaped)
		properties++
	}
	if embed.Description != "" {
		size += len("description") + 5 + jsonStringSize(embed.Description, &escaped)
		properties++
	}
	if embed.URL != "" {
		size += len("url") + 5 + jsonStringSize(embed.URL, &escaped)
		properties++
	}
	if embed.Timestamp != "" {
		size += len("timestamp") + 5 + jsonStringSize(embed.Timestamp, &escaped)
		properties++
	}
	if embed.Color != nil {
		size += len("color") + 3 + 8
		properties++
	}

	if embed.Footer != nil {
		footerSize := 2 + len("text") + 5 + jsonStringSize(embed.Footer.Text, &escaped)
		if embed.Footer.IconURL != "" {
			footerSize += 1 + len("icon_url") + 5 + jsonStringSize(embed.Footer.IconURL, &escaped)
		}

		size += len("footer") + 3 + footerSize
		properties++
	}
	if embed.Image != nil {
		mediaSize := 2 + len("url") + 5 + jsonStringSize(embed.Image.URL, &escaped)
		size += len("image") + 3 + mediaSize
		properties++
	}
	if embed.Thumbnail != nil {
		mediaSize := 2 + len("url") + 5 + jsonStringSize(embed.Thumbnail.URL, &escaped)
		size += len("thumbnail") + 3 + mediaSize
		properties++
	}
	if embed.Author != nil {
		authorSize := 2 + len("name") + 5 + jsonStringSize(embed.Author.Name, &escaped)
		if embed.Author.URL != "" {
			authorSize += 1 + len("url") + 5 + jsonStringSize(embed.Author.URL, &escaped)
		}
		if embed.Author.IconURL != "" {
			authorSize += 1 + len("icon_url") + 5 + jsonStringSize(embed.Author.IconURL, &escaped)
		}

		size += len("author") + 3 + authorSize
		properties++
	}

	if len(embed.Fields) != 0 {
		fieldsSize := 2 + len(embed.Fields) - 1
		for _, field := range embed.Fields {
			fieldSize := 2 + len("name") + 5 + jsonStringSize(field.Name, &escaped)
			fieldSize += 1 + len("value") + 5 + jsonStringSize(field.Value, &escaped)
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

	return size, escaped
}
