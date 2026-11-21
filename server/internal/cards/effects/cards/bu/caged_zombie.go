package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Caged Zombie — Creature — Zombie {2}{B}, 2/3 (Core Set 2021, #5511):
//
//	"{1}{B}, {T}: Each opponent loses 2 life. Activate only if a creature
//	 died this turn."
//
// Undercity Scrounger's gate (ACreatureDiedThisTurn, the table-wide
// tally of creatures that died, tokens included) on a drain of each
// opponent. The {T} is the creature's own, so a summoning-sick Zombie
// can't use it (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "430b3e61-09a5-4390-b377-89830cacf66e",
		Name:         "Caged Zombie",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{1}{B}, {T}: Each opponent loses 2 life. Activate only if a creature died this turn.",
			Purpose:   game.Purpose{Answers: game.AnswerValue},
			Cost:      effects.Plus(effects.ManaCost("{1}{B}"), effects.TapCost()),
			Condition: effects.ACreatureDiedThisTurn(),
			Effect:    effects.Do(effects.EachOpponentLosesLife{N: 2}),
		}},
	})
}
