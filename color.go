package embedbuilder

import (
	"errors"
)

// ErrInvalidHexColor identifies a malformed hexadecimal color.
var ErrInvalidHexColor = errors.New("embedbuilder: invalid hexadecimal color")

// SetColorRGB sets the embed accent color from red, green, and blue channels.
func (b *Builder) SetColorRGB(red, green, blue uint8) *Builder {
	color := uint32(red)<<16 | uint32(green)<<8 | uint32(blue)

	return b.SetColor(color)
}

// SetColorHex sets the embed accent color from RGB, #RGB, RRGGBB, or #RRGGBB. Three-digit colors expand each digit, so
// 321 becomes 332211. The builder is unchanged if color is invalid.
func (b *Builder) SetColorHex(color string) error {
	value := color
	if len(value) != 0 && value[0] == '#' {
		value = value[1:]
	}

	short := len(value) == 3
	if !short && len(value) != 6 {
		return ErrInvalidHexColor
	}

	var parsedColor uint32
	if short {
		for index := 0; index < len(value); index++ {
			digit, valid := hexadecimalDigit(value[index])
			if !valid {
				return ErrInvalidHexColor
			}

			parsedColor = parsedColor<<8 | digit*0x11
		}
	} else {
		for index := 0; index < len(value); index++ {
			digit, valid := hexadecimalDigit(value[index])
			if !valid {
				return ErrInvalidHexColor
			}

			parsedColor = parsedColor<<4 | digit
		}
	}

	b.SetColor(parsedColor)

	return nil
}

func hexadecimalDigit(character byte) (uint32, bool) {
	switch {
	case character >= '0' && character <= '9':
		return uint32(character - '0'), true
	case character >= 'a' && character <= 'f':
		return uint32(character-'a') + 10, true
	case character >= 'A' && character <= 'F':
		return uint32(character-'A') + 10, true
	default:
		return 0, false
	}
}
