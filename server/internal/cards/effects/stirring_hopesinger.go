package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Stirring Hopesinger — Creature — Bird Bard {2}{W}, 1/3 (Secrets of
// Strixhaven, #4456):
//
//	"Flying, lifelink
//	 Repartee — Whenever you cast an instant or sorcery spell that targets a
//	 creature, put a +1/+1 counter on each creature you control."
//
// Repartee is Informed Inkwright's trigger (informedInkwrightCast: the cast
// spell's own announced targets). The creatures that get a counter are read
// when the trigger resolves, as Cathars' Crusade does, so the Hopesinger
// itself counts.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "7244d04a-bb9c-4fbe-a346-4a3f45709da7",
		Name:            "Stirring Hopesinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "lifelink"},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, informedInkwrightCast,
				"Stirring Hopesinger — Repartee: a +1/+1 counter on each creature you control",
				func(g *game.Game, item *game.StackItem) error {
					return b11PutCountersOnEachCreatureYouControl(g, item, 1)
				}),
		},
	})
}
