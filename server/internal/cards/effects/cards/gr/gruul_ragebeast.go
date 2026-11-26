package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gruul Ragebeast — Creature — Beast {5}{R}{G}, 6/6 (Gatecrash, #5565):
//
//	"Whenever this creature or another creature you control enters, that
//	 creature fights target creature an opponent controls."
//
// Thorn Mammoth's one trigger on every creature entering under your control,
// itself included, with the fighter being the creature that ENTERED rather
// than the Ragebeast, and a required target (so with no creature an opponent
// controls nothing triggers). Each creature entering is its own trigger
// (CR 603.2c). The entering creature must still be on the battlefield and
// the target still legal as it resolves, else nobody fights (CR 701.14b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5fa93a3-0b2b-4477-ad3f-4fc1b325ca19",
		Name:         "Gruul Ragebeast",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(
				effects.On(game.EventETB, effects.CreatureEnteredUnderYourControl,
					"Gruul Ragebeast — that creature fights target creature an opponent controls",
					func(g *game.Game, item *game.StackItem) error {
						if item.Trigger == nil {
							return nil
						}
						ctx := effects.NewContext(g, item)
						id, ok := effects.B16FirstLegalTargetCard(ctx)
						if !ok {
							return nil
						}
						return effects.B10Fight(ctx, item.Trigger.Event.CardID, id)
					}),
				effects.TargetCreature("target creature an opponent controls", effects.OpponentControls())),
		},
	})
}
