package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hollow One — Artifact Creature — Golem {5}, 4/4 (Hour of Devastation;
// Modern Horizons 3 and others, #5948):
//
//	"This spell costs {2} less to cast for each card you've cycled or
//	 discarded this turn.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The count is the caster's per-turn discard tally
// (game.PlayerTurnTally.CardsDiscarded, one per card discarded, whatever the
// cause). Cycling discards the card as its cost (CR 702.29a), so a cycled
// card is already among the discards and counts once, not twice — the
// card's ruling. The reduction is generic mana only (CR 601.2f), so it can
// take the spell to {0} and no lower.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b9e4d34a-ba91-4ad6-b886-daeeda6c9498",
		Name:         "Hollow One",
		Completeness: effects.CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessEach(func(q game.CostQuery) int {
				if q.Game == nil {
					return 0
				}
				return 2 * q.Game.TurnTallyFor(q.Controller).CardsDiscarded
			}, "This spell costs {2} less to cast for each card you've cycled or discarded this turn."),
		},
		Activated: []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
