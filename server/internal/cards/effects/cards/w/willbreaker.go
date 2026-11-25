package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Willbreaker — Creature — Human Wizard {3}{U}{U}, 2/3 (Magic Origins, S82 older
// sets):
//
//	"Whenever a creature an opponent controls becomes the target of a spell or
//	 ability you control, gain control of that creature for as long as you
//	 control this creature."
//
// Shay Cormac's becomes-the-target trigger (one per target slot, CR 115.3, for a
// creature an opponent of Willbreaker's controller controls and a spell or
// ability its controller controls) with Mind Flayer's duration: CR 611.2b's "for
// as long as you control ~", so the creature goes home when Willbreaker leaves
// or changes hands, and nothing is taken if it already has. Willbreaker's own
// controller's creatures never trigger it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7c365041-806f-42d6-a2ff-40e20dd8e9f0",
		Name:         "Willbreaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventBecomesTarget, effects.OldB322CreatureAnOpponentControlsBecameYourTarget, "Willbreaker — gain control of that creature for as long as you control this creature",
				func(g *game.Game, item *game.StackItem) error {
					id := item.Trigger.Event.CardID
					if !effects.OnBattlefield(g, id) {
						return nil
					}
					ctx := effects.NewContext(g, item)
					d, ok := effects.DurationWhileYouControlSource(ctx, item.SourceCardID, item.Controller)
					if !ok {
						return nil
					}
					return effects.GainControl{
						Target:     id,
						Controller: item.Controller,
						Duration:   d,
						Label:      "Willbreaker — control for as long as you control Willbreaker",
					}.Apply(ctx)
				}),
		},
	})
}
