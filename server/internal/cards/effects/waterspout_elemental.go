package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Waterspout Elemental — Creature — Elemental {3}{U}{U}, 3/4:
//
//	"Kicker {U}
//	 Flying
//	 When this creature enters, if it was kicked, return all other
//	 creatures to their owners' hands and you skip your next turn."
//
// An intervening if (CR 603.4) read off the permanent's carried kicker
// record when the ability would trigger, so an unkicked Elemental puts
// nothing on the stack. The mass bounce is symmetric and spares only
// this creature; the skip follows it (CR 614.10, ADR 0188 Decision 1,
// #3565).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f86b8c03-9f1e-4d42-ac45-907c46e9e1c0",
		Name:            "Waterspout Elemental",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		OptionalCosts:   []game.AdditionalCost{Kicker("{U}")},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return game.CardKickedTimes(*source) > 0
			}), "Waterspout Elemental — return all other creatures to their owners' hands and you skip your next turn",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (BounceAllMatching{Match: And(Creature(), NotSelf(ctx.Source()))}).Apply(ctx); err != nil {
						return err
					}
					return SkipYourNextTurns{}.Apply(ctx)
				}),
		},
	})
}
