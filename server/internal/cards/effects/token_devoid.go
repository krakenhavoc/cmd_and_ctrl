package effects

import (
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// dropKeyword removes kw (case-insensitive) from the token's keyword list.
// CR 707.9d: when a copy effect provides a value for a characteristic, the
// copied object's CDA that would define that characteristic is not copied —
// a colour-setting exception therefore must not leave devoid on a coloured
// token (#2322).
func dropKeyword(t *game.Card, kw string) {
	if t == nil {
		return
	}
	out := t.Keywords[:0]
	for _, k := range t.Keywords {
		if !strings.EqualFold(k, kw) {
			out = append(out, k)
		}
	}
	for i := len(out); i < len(t.Keywords); i++ {
		t.Keywords[i] = ""
	}
	t.Keywords = out
}
