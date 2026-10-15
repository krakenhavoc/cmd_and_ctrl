package boardtext

import (
	"strconv"
	"strings"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/protocol"
)

// manaPoolLettersMax is the most mana manaPoolText spells out one
// letter at a time.
const manaPoolLettersMax = 20

// manaPoolText renders a seat's floating pool the way it always read,
// one letter per mana in pool order ("GGR"), unless the pool is too
// big to count by eye (#3041: a million {G} is one tap of a big
// Marwyn). Then each run reads as its count and letter, "1048576 G,
// 2 R".
func manaPoolText(pool []protocol.ManaPoolRun) string {
	if protocol.ManaPoolTotal(pool) <= manaPoolLettersMax {
		var b strings.Builder
		for _, r := range pool {
			b.WriteString(strings.Repeat(r.Color, r.Count))
		}
		return b.String()
	}
	parts := make([]string, 0, len(pool))
	for _, r := range pool {
		parts = append(parts, strconv.Itoa(r.Count)+" "+r.Color)
	}
	return strings.Join(parts, ", ")
}
