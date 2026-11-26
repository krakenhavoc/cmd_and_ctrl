package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Essence Symbiote — Creature — Beast {1}{G}, 2/2 (Ikoria: Lair of Behemoths, #5512):
//
//	"Whenever a creature you control mutates, put a +1/+1 counter on that
//	 creature and you gain 2 life."
//
// The creature that mutated is the permanent a mutating spell merged with
// (CR 730.2c), under your control when it did, the Symbiote itself included.
// The counter goes on it if it is still on the battlefield as the trigger
// resolves; the life is gained either way (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9875cb1b-d67e-4ac4-8c16-164f28d4e4d4",
		Name:         "Essence Symbiote",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventMutate, effects.S82OldB114MutatedCreatureYouControl,
				"Essence Symbiote — put a +1/+1 counter on that creature and gain 2 life",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := item.Trigger.Event.CardID
					if effects.OnBattlefield(g, id) {
						if err := (effects.AddCounter{Target: id, Kind: game.CounterPlusOne, N: 1}).Apply(ctx); err != nil {
							return err
						}
					}
					return effects.GainLife{Amount: 2}.Apply(ctx)
				}),
		},
	})
}
