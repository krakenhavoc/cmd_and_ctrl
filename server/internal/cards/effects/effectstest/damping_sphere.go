package effectstest

import (
	"sort"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

func SortedColors(p *game.Player) []string {
	out := Batch01PoolColors(p)
	sort.Strings(out)
	return out
}
