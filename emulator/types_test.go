package emulator

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWord(t *testing.T) {
	t.Run("test word", func(t *testing.T) {
		w := NewWord(0x1F, 0x90)
		assert.Equal(t, Word(0x1F90), w)
		assert.Equal(t, uint8(0x1F), w.High())
		assert.Equal(t, uint8(0x90), w.Low())
	})
}
