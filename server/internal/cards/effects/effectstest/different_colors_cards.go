package effectstest

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func DcSortedPool(p *game.Player) []string {
	out := make([]string, 0, p.ManaPool.Total())
	for _, tok := range p.ManaPool.Units() {
		out = append(out, tok.Color)
	}
	sort.Strings(out)
	return out
}
