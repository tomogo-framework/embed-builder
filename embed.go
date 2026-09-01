// Package embedbuilder provides a dependency-free, fluent builder for Discord
// rich embeds.
package embedbuilder

// Discord embed limits. Text limits are measured in Unicode characters after leading and trailing whitespace is
// removed, matching Discord's documented behavior.
const (
	MaxTitleCharacters       = 256
	MaxDescriptionCharacters = 4096
	MaxFields                = 25
	MaxFieldNameCharacters   = 256
	MaxFieldValueCharacters  = 1024
	MaxFooterCharacters      = 2048
	MaxAuthorNameCharacters  = 256
	MaxTotalCharacters       = 6000
	MaxEmbedsPerMessage      = 10
	MaxColor                 = 0xFFFFFF
)

// Generic colors.
const (
	ColorBlack uint32 = 0x000000
	ColorWhite uint32 = MaxColor
)

// Official Discord brand colors.
const (
	ColorDark         uint32 = 0x1F1F1F
	ColorBlurple      uint32 = 0x5865F2
	ColorLightBlurple uint32 = 0xE0E3FF
	ColorDarkBlurple  uint32 = 0x19175C
	ColorGreen        uint32 = 0x35ED7E
	ColorLightGreen   uint32 = 0xC8FFEF
	ColorDarkGreen    uint32 = 0x002920
	ColorPink         uint32 = 0xFF4CD2
	ColorLightPink    uint32 = 0xF5C9FF
	ColorDarkPink     uint32 = 0x381F2C
)

// Embed is a framework-agnostic representation of a Discord rich embed. It can be passed to any Discord client or
// marshaled directly with encoding/json.
type Embed struct {
	Title       string  `json:"title,omitempty"`
	Description string  `json:"description,omitempty"`
	URL         string  `json:"url,omitempty"`
	Timestamp   string  `json:"timestamp,omitempty"`
	Color       *uint32 `json:"color,omitempty"`
	Footer      *Footer `json:"footer,omitempty"`
	Image       *Media  `json:"image,omitempty"`
	Thumbnail   *Media  `json:"thumbnail,omitempty"`
	Author      *Author `json:"author,omitempty"`
	Fields      []Field `json:"fields,omitempty"`
}

// Author contains the author information displayed above an embed.
type Author struct {
	Name    string `json:"name"`
	URL     string `json:"url,omitempty"`
	IconURL string `json:"icon_url,omitempty"`
}

// Footer contains the text and optional icon displayed below an embed.
type Footer struct {
	Text    string `json:"text"`
	IconURL string `json:"icon_url,omitempty"`
}

// Media identifies an embed image or thumbnail.
type Media struct {
	URL string `json:"url"`
}

// Field is a named value displayed in an embed.
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Inline bool   `json:"inline,omitempty"`
}
