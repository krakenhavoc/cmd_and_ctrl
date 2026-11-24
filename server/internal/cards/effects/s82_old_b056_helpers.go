package effects

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b056_helpers.go — shared bodies for OLD batch 56 (#5504).

// AnyPlayerCastAnInstantOrSorcery is "whenever a player casts an instant or
// sorcery spell" — the spell is read off the stack, and the caster is
// anyone (Niv-Mizzet, Parun; Geometric Nexus).
func AnyPlayerCastAnInstantOrSorcery(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
	if ev.Kind != game.EventCast {
		return false
	}
	spell, ok := g.LookupCardForEffect(ev.CardID)
	return ok && (spell.IsInstant() || spell.IsSorcery())
}
