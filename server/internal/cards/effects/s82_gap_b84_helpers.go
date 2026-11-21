package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_gap_b84_helpers.go — shared bodies for the GAP set, batch 84 (Doctor
// Who creatures).

// gapB84FirstNonlegendaryArtifactSpell is "the first nonlegendary artifact
// spell you cast each turn", read as the spell is announced, before the cast
// is tallied (CR 601.2b): the spell is a nonlegendary artifact and the caster
// has cast no nonlegendary artifact spell yet this turn (a legendary artifact
// spell cast earlier does not count).
func gapB84FirstNonlegendaryArtifactSpell(g *game.Game, caster uuid.UUID, spell game.Card) bool {
	if !spell.IsArtifact() || spell.IsLegendary() {
		return false
	}
	for _, ev := range g.EventsThisTurn() {
		if ev.Kind != game.EventCast || ev.Actor != caster || ev.CardID == spell.InstanceID {
			continue
		}
		prior, ok := g.LookupCardForEffect(ev.CardID)
		if ok && prior.IsArtifact() && !prior.IsLegendary() {
			return false
		}
	}
	return true
}
