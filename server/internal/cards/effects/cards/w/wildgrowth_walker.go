package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wildgrowth Walker — Creature — Elemental {1}{G}, 1/3 (Ixalan, #5532):
//
//	"Whenever a creature you control explores, put a +1/+1 counter on this
//	 creature and you gain 3 life."
//
// Merfolk Cave-Diver's trigger (CR 701.44b): one for each explore a creature
// you control completes, the Walker's own included. The counter goes on only
// while this is still the same permanent; the life is gained either way.
//
// No simplification.
func init() {
	const label = "Wildgrowth Walker — put a +1/+1 counter on this creature and you gain 3 life"
	effects.Register(effects.Spec{
		OracleID:     "d80562c2-cf90-43a4-9721-28da40fdfeed",
		Name:         "Wildgrowth Walker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventExplored, effects.ACreatureYouControlExplored, label,
				func(g *game.Game, item *game.StackItem) error {
					if err := effects.PlusOneCountersOnThis(1)(g, item); err != nil {
						return err
					}
					return effects.GainLife{Amount: 3}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}
