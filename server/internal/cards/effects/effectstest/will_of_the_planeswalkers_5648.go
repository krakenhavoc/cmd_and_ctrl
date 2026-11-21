package effectstest

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// --- Ichor Elixir ----------------------------------------------------------------

// planarRollsSince lists the faces of the planar rolls announced after
// the first `from` events.
func PlanarRollsSince(g *game.Game, from int) []string {
	var out []string
	for _, ev := range g.Events[from:] {
		if ev.Kind == game.EventRollDie && ev.PlanarFace != "" {
			out = append(out, ev.PlanarFace)
		}
	}
	return out
}
