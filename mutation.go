package embedbuilder

import (
	"errors"
	"fmt"
)

// ErrFieldIndexOutOfRange identifies an invalid field mutation index.
var ErrFieldIndexOutOfRange = errors.New("embedbuilder: field index out of range")

// CharacterCount returns the number of characters counted toward Discord's per-message embed text limit.
func (b *Builder) CharacterCount() int {
	if b == nil {
		return 0
	}

	return b.counts.total
}

// RemainingCharacters returns the remaining text capacity, clamped to zero.
func (b *Builder) RemainingCharacters() int {
	remaining := MaxTotalCharacters - b.CharacterCount()
	if remaining < 0 {
		return 0
	}

	return remaining
}

// FieldCount returns the number of fields currently in the builder.
func (b *Builder) FieldCount() int {
	if b == nil {
		return 0
	}

	return len(b.embed.Fields)
}

// CanAddField reports whether adding the field would preserve every current validation rule and Discord size limit.
func (b *Builder) CanAddField(name, value string) bool {
	if b == nil || len(b.embed.Fields) >= MaxFields || b.Validate() != nil {
		return false
	}

	nameCharacters := characterCount(name)
	if nameCharacters == 0 || nameCharacters > MaxFieldNameCharacters {
		return false
	}

	valueCharacters := characterCount(value)
	if valueCharacters == 0 || valueCharacters > MaxFieldValueCharacters {
		return false
	}

	return b.counts.total+nameCharacters+valueCharacters <= MaxTotalCharacters
}

// SetField replaces a field at index.
func (b *Builder) SetField(index int, name, value string, inline bool) error {
	if err := b.validateFieldIndex(index, false); err != nil {
		return err
	}

	oldCharacters := characterCount(b.embed.Fields[index].Name) + characterCount(b.embed.Fields[index].Value)
	nameCharacters := characterCount(name)
	valueCharacters := characterCount(value)
	newCharacters := nameCharacters + valueCharacters

	b.embed.Fields[index] = Field{Name: name, Value: value, Inline: inline}
	b.counts.fields += newCharacters - oldCharacters
	b.counts.total += newCharacters - oldCharacters
	b.counts.rulesDirty = true
	if len(b.counts.fieldViolations) != 0 || nameCharacters > MaxFieldNameCharacters || valueCharacters > MaxFieldValueCharacters {
		b.recalculateFields()
	}

	return nil
}

// InsertField inserts a field before index. An index equal to FieldCount is valid and appends the field.
func (b *Builder) InsertField(index int, name, value string, inline bool) error {
	if err := b.validateFieldIndex(index, true); err != nil {
		return err
	}

	b.ensureFieldCapacity()
	b.embed.Fields = append(b.embed.Fields, Field{})
	copy(b.embed.Fields[index+1:], b.embed.Fields[index:])
	b.embed.Fields[index] = Field{Name: name, Value: value, Inline: inline}
	nameCharacters := characterCount(name)
	valueCharacters := characterCount(value)
	fieldCharacters := nameCharacters + valueCharacters
	b.counts.fields += fieldCharacters
	b.counts.total += fieldCharacters
	b.counts.rulesDirty = true
	if len(b.counts.fieldViolations) != 0 || nameCharacters > MaxFieldNameCharacters || valueCharacters > MaxFieldValueCharacters {
		b.recalculateFields()
	}

	return nil
}

// RemoveField removes a field at index.
func (b *Builder) RemoveField(index int) error {
	if err := b.validateFieldIndex(index, false); err != nil {
		return err
	}

	removedCharacters := characterCount(b.embed.Fields[index].Name) + characterCount(b.embed.Fields[index].Value)

	copy(b.embed.Fields[index:], b.embed.Fields[index+1:])
	last := len(b.embed.Fields) - 1
	b.embed.Fields[last] = Field{}
	b.embed.Fields = b.embed.Fields[:last]
	b.counts.fields -= removedCharacters
	b.counts.total -= removedCharacters
	b.counts.rulesDirty = true
	if len(b.counts.fieldViolations) != 0 {
		b.recalculateFields()
	}

	return nil
}

func (b *Builder) validateFieldIndex(index int, allowEnd bool) error {
	if b == nil {
		return ErrNilBuilder
	}

	limit := len(b.embed.Fields)
	if allowEnd {
		limit++
	}
	if index < 0 || index >= limit {
		return fmt.Errorf("%w: index %d with %d fields", ErrFieldIndexOutOfRange, index, len(b.embed.Fields))
	}

	return nil
}

func (b *Builder) recalculateFields() {
	b.counts.total -= b.counts.fields
	b.counts.fields = 0
	b.counts.fieldViolations = b.counts.fieldViolations[:0]

	for index, field := range b.embed.Fields {
		nameCharacters := characterCount(field.Name)
		valueCharacters := characterCount(field.Value)
		b.counts.fields += nameCharacters + valueCharacters

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

	b.counts.total += b.counts.fields
	b.counts.rulesDirty = true
}
