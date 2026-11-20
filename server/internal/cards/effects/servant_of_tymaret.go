package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Servant of Tymaret — Creature — Zombie {2}{B}, 1/3 (Born of the Gods,
// #5559):
//
//	"Inspired — Whenever this creature becomes untapped, each opponent
//	 loses 1 life. You gain life equal to the life lost this way.
//	 {2}{B}: Regenerate this creature."
//
// Inspired is an ability word: the trigger is Key to the City's
// (EventUntapCard on itself), so it fires in the untap step and on
// any other untap.
// The drain gains the life actually lost
// (b21DrainEachOpponentAndGainTheTotal), and the regeneration is the
// shared row (regenerateThisRow).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1ea5a8c2-d8c4-4028-b1fe-9951514ea882",
		Name:         "Servant of Tymaret",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventUntapCard, Self,
				"Servant of Tymaret — each opponent loses 1 life and you gain the life lost (inspired)",
				func(g *game.Game, item *game.StackItem) error {
					return b21DrainEachOpponentAndGainTheTotal(g, item, 1)
				}),
		},
		Activated: []ActivatedAbility{regenerateThisRow("{2}{B}")},
	})
}
