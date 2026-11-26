package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b475_helpers.go — shared pieces of the Magic 2011 damage cards,
// S82 older-set batch 475 (#5579).

// S82OldB475ControlledByDefender is "target creature defending player
// controls" on an ACTIVATED ability of an attacking creature (Ancient
// Hellkite): the candidate is controlled by the player the source is
// attacking (game.CombatControlledByDefender, CR 506.3d). It is judged at
// announce and again at resolution, and a source that is not attacking has
// no such creature to target. A trigger of an attacker uses
// TargetCreatureDefendingPlayerControls instead.
func S82OldB475ControlledByDefender(spec *game.TargetSpec) *game.TargetSpec {
	spec.CombatWithSource = game.SourceCombat{Of: game.CombatOfSource, Role: game.CombatControlledByDefender}
	return spec
}

// S82OldB475SeatsOwing lists every living seat n times: a name listed N
// times owes N permanents (PlayersSacrificeThen).
func S82OldB475SeatsOwing(g *game.Game, n int) []uuid.UUID {
	var out []uuid.UUID
	for _, seat := range ApnapLivingSeats(g) {
		for i := 0; i < n; i++ {
			out = append(out, seat)
		}
	}
	return out
}
