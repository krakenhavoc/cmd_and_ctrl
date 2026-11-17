package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Chittering Skitterling — Creature — Phyrexian Rat {2}{B}, 1/4:
//
//	"Corrupted — Sacrifice an artifact or creature: Draw a card. Activate
//	 only if an opponent has three or more poison counters and only once
//	 each turn."
//
// The sacrifice is a cost (CR 602.2b), so the Skitterling is itself a legal
// sacrifice. Both limits are activation conditions read as it is announced:
// the poison, and the once-per-turn tally, which counts announcements.
//
// No simplification.
func init() {
	const label = "Corrupted — Sacrifice an artifact or creature: Draw a card. Activate only if an opponent has three or more poison counters and only once each turn."
	Register(Spec{
		OracleID:     "864fc52a-e9f4-4254-901e-b949d96b34eb",
		Name:         "Chittering Skitterling",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Draws: 1},
			Cost:      SacrificeN(1, "an artifact or creature", Or(Artifact(), Creature())),
			Condition: corruptedAnd(OncePerTurnActivation(label)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return Do(DrawCards{N: 1})(g, item)
			},
		}},
	})
}
