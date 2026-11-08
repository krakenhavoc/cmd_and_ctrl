package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Heartfire Hero — Creature — Mouse Soldier {R}, 1/1 (Bloomburrow, #5230):
//
//	"Valiant — Whenever this creature becomes the target of a spell or
//	 ability you control for the first time each turn, put a +1/+1 counter
//	 on it.
//	 When this creature dies, it deals damage equal to its power to each
//	 opponent."
//
// Valiant is the shared trigger (valiant.go); the counter goes on only
// while the Hero is still the permanent that was targeted (CounterOnThis).
// The dies trigger is Juri, Master of the Revue's: the power is the Hero's
// as it last existed on the battlefield (CR 608.2h), counters and pumps
// included, read off the trigger's record at resolution, and the damage is
// dealt by the Hero to each opponent at once (Origin of Thor's shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "49e1e202-37bf-46ed-93d1-1adb13726039",
		Name:         "Heartfire Hero",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			Valiant("Heartfire Hero — Valiant: put a +1/+1 counter on it",
				Do(CounterOnThis{Kind: game.CounterPlusOne, N: 1})),
			WhenThisDies("Heartfire Hero — it deals damage equal to its power to each opponent",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					info, _ := ctx.TriggeringPermanent()
					if info.Power <= 0 {
						return nil
					}
					return originOfThorDealToEachOpponent(ctx, item.SourceCardID, info.Power)
				}),
		},
	})
}
