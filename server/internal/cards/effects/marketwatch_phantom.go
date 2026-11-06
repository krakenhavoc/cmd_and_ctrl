package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Marketwatch Phantom — Creature — Spirit Detective {1}{W}, 2/2
// (Murders at Karlov Manor, #4469):
//
//	"Whenever another creature you control with power 2 or less enters,
//	 this creature gains flying until end of turn."
//
// Mentor of the Meek's trigger condition, read as the creature enters.
// The flying is pinned to the Phantom; one that left the battlefield
// before the trigger resolved gets nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6b0d4192-bc76-4ef1-92b2-b5fca81970bd",
		Name:         "Marketwatch Phantom",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AnotherCreatureWithPower2OrLessEnteredUnderYourControl,
				"Marketwatch Phantom — this creature gains flying until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					if !onBattlefield(g, item.SourceCardID) {
						return nil
					}
					return GrantKeywordUntilEOT{
						Target:   item.SourceCardID,
						Keywords: []string{"flying"},
						Label:    "Marketwatch Phantom — flying until end of turn",
					}.Apply(NewContext(g, item))
				}),
		},
	})
}
