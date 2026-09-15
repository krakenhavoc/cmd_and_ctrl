package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Corpse Knight — Creature — Zombie Knight {W}{B}, 2/2 (EDHREC rank
// 1427):
//
//	"Whenever another creature you control enters, each opponent
//	 loses 1 life."
//
// The Orzhov token deck's drain. One trigger per creature entering
// — a mass token creation fires it once per token, which is the
// printed total — and life LOSS, not damage, so no prevention or
// damage doubler sees it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0c77b805-728d-4ae4-88a9-223f4b7b52e0",
		Name:         "Corpse Knight",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverAnotherCreatureEntersUnderYourControl("Corpse Knight — each opponent loses 1 life", func(g *game.Game, item *game.StackItem) error {
				return eachOpponentLosesLife(g, item, 1)
			}),
		},
	})
}
