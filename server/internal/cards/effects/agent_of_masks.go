package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Agent of Masks — Creature — Human Advisor {3}{W}{B}, 2/3 (Guildpact,
// #5599):
//
//	"At the beginning of your upkeep, each opponent loses 1 life. You
//	 gain life equal to the life lost this way."
//
// Gray Merchant's drain at a fixed 1 (b21DrainEachOpponentAndGainTheTotal):
// life loss, not damage, and the gain is what the opponents really lost
// once every loss has settled, so an opponent whose loss is replaced
// away adds nothing to it.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "632ccf7e-eaf4-42a6-9069-6ad915d03a5b",
		Name:         "Agent of Masks",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			AtYourUpkeep("Agent of Masks — each opponent loses 1 life and you gain the life lost",
				func(g *game.Game, item *game.StackItem) error {
					return b21DrainEachOpponentAndGainTheTotal(g, item, 1)
				}),
		},
	})
}
