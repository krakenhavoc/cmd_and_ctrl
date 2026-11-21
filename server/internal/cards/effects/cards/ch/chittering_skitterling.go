package cardsch

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "864fc52a-e9f4-4254-901e-b949d96b34eb",
		Name:         "Chittering Skitterling",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     label,
			Purpose:   game.Purpose{Draws: 1, Answers: game.AnswerSacOutlet},
			Cost:      effects.SacrificeN(1, "an artifact or creature", effects.Or(effects.Artifact(), effects.Creature())),
			Condition: effects.CorruptedAnd(effects.OncePerTurnActivation(label)),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.Do(effects.DrawCards{N: 1})(g, item)
			},
		}},
	})
}
