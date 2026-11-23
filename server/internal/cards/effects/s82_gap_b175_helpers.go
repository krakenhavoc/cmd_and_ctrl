package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// TargetsACreatureOrPlaneswalkerYouControl passes when one of the spell's
// announced targets is a creature or planeswalker the modifier source's
// controller controls — Kasmina, Enigmatic Mentor's "Spells your
// opponents cast that target a creature or planeswalker you control cost
// {2} more to cast". Pair it with OpponentsSpell and set ReadsTargets.
func TargetsACreatureOrPlaneswalkerYouControl() CostPredicate {
	return func(q game.CostQuery) bool {
		if q.Game == nil {
			return false
		}
		for _, t := range q.Targets {
			if t.Kind != game.TargetCard {
				continue
			}
			if c, ok := q.Game.LookupCardForEffect(t.ID); ok && c.Controller == q.Source.Controller &&
				(c.IsCreature() || c.IsPlaneswalker()) {
				return true
			}
		}
		return false
	}
}

// YouHaveAtLeastLifeAboveStarting reports whether `player` has at least n
// life more than the table's starting life total (ADR 0075) — Ajani,
// Strength of the Pride's 0. Read as the ability resolves.
func YouHaveAtLeastLifeAboveStarting(g *game.Game, player uuid.UUID, n int) bool {
	p := g.PlayerByIDForEffect(player)
	return p != nil && p.Life >= StartingLifeOf(g)+n
}
