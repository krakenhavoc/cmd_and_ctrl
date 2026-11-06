package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Water Whip — Sorcery — Lesson {U}{U} (Avatar: The Last Airbender
// Eternal, #5075):
//
//	"As an additional cost to cast this spell, waterbend {5}. (While
//	 paying a waterbend cost, you can tap your artifacts and creatures to
//	 help. Each one pays for {1}.)
//	 Return up to two target creatures to their owners' hands. Draw two
//	 cards."
//
// The waterbend is Benevolent River Spirit's additional cost. The two
// creatures go back in one bounce (Horses of the Bruinen's), then the
// draw waits on any prompt it asked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2015f455-988e-444c-93df-23fab179ef87",
		Name:         "Water Whip",
		Completeness: CompletenessFull,
		TapCost:      Waterbend("{5}"),
		Targets:      TargetCreature("up to two target creatures").WithCount(0, 2),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ctx.Game.BounceCardsToHandKeyedForEffect(legalTargetIDs(ctx), ctx.TailN(avB34ThenDraw, 2))
		},
	})
}
