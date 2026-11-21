package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/legal"
)

// castMove reports whether the legal-move enumerator offers casting
// `spell` — the list the ready ring and every bot read.
func CastMove(g *game.Game, seat, spell uuid.UUID) bool {
	for _, m := range legal.EnumerateFor(g, seat) {
		if m.Type == "cast_spell" && m.Source == spell {
			return true
		}
	}
	return false
}
