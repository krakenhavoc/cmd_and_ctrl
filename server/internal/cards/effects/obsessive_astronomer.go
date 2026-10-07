package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Obsessive Astronomer — {1}{R} Creature — Human Wizard 2/2 (#2561, ADR
// 0132):
//
//	"If it's neither day nor night, it becomes day as this creature enters.
//	 Whenever day becomes night or night becomes day, discard up to two
//	 cards, then draw that many cards."
//
// The same discard-then-draw run Fable of the Mirror-Breaker's chapter II
// uses: the count drawn is the count actually discarded.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "fde06097-1cf1-4774-b6ca-7cf15dfe4837",
		Name:         "Obsessive Astronomer",
		Completeness: CompletenessFull,
		AsEnters:     BecomesDayAsEnters(),
		Triggered: []game.TriggeredAbility{
			WheneverDayBecomesNightOrNightBecomesDay("Obsessive Astronomer — discard up to two cards, then draw that many",
				func(g *game.Game, item *game.StackItem) error {
					return b39MayDiscardThenDraw(2, false,
						"Obsessive Astronomer — discard up to two cards, then draw that many",
						func(discarded int) int { return discarded })(NewContext(g, item))
				}),
		},
	})
}
