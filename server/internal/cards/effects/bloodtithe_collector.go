package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bloodtithe Collector — Creature — Vampire Noble {4}{B}, 3/4
// (Foundations, #4465):
//
//	"Flying
//	 When this creature enters, if an opponent lost life this turn, each
//	 opponent discards a card."
//
// Lion Vulture's "if an opponent lost life this turn" on Burglar Rat's
// enters discard. The condition is an intervening "if" (CR 603.4):
// checked as the trigger would fire and again as it resolves. Each
// opponent chooses the card they discard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "93088b0c-d158-473a-9c06-aa2f344a9868",
		Name:            "Bloodtithe Collector",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			On(game.EventETB, AllOf(Self, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return opponentLifeLostThisTurn(g, source.Controller) > 0
			}), "Bloodtithe Collector — each opponent discards a card (an opponent lost life this turn)",
				Do(If{
					Cond: avB34CondAnOpponentLostLife{},
					Then: []Applier{Instruction(func(ctx *Context) error { return eachOpponentDiscardsOne(ctx.Game, ctx.Item) })},
				})),
		},
	})
}
