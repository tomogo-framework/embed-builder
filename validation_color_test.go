package embedbuilder

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestColorValidationAcrossIntegerWidths(t *testing.T) {
	t.Parallel()

	for _, color := range []uint32{0, MaxColor, MaxColor + 1, 0x7FFFFFFF, 0x80000000, 0xFFFFFFFF} {
		t.Run(fmt.Sprintf("%08x", color), func(t *testing.T) {
			t.Parallel()

			builder := New().SetTitle("color").SetColor(color)
			_, buildError := builder.Build()
			_, jsonError := builder.BuildJSON()
			checks := []error{
				builder.Validate(),
				buildError,
				jsonError,
				ValidateEmbed(builder.embed),
				ValidateEmbeds(builder.embed),
				NewCollection().Add(builder.embed).Validate(),
			}
			for index, err := range checks {
				if color <= MaxColor {
					if err != nil {
						t.Errorf("check %d rejected valid color: %v", index, err)
					}
					continue
				}

				validationError := requireValidationError(t, err)
				if len(validationError.Violations) != 1 {
					t.Fatalf("check %d: unexpected violations: %#v", index, validationError.Violations)
				}

				wantMessage := fmt.Sprintf("color has %d, limit is %d", color, MaxColor)
				if !strings.Contains(err.Error(), wantMessage) {
					t.Errorf("check %d: error = %q, want %q", index, err, wantMessage)
				}

				violation := validationError.Violations[0]
				if strconv.IntSize == 32 && color > 1<<31-1 {
					if violation.Rule == "" || violation.Actual != 0 || violation.Limit != 0 {
						t.Errorf("check %d: unrepresentable value must use Rule: %#v", index, violation)
					}
				} else if uint64(violation.Actual) != uint64(color) || violation.Limit != MaxColor {
					t.Errorf("check %d: unexpected numeric limits: %#v", index, violation)
				}
			}
		})
	}
}
