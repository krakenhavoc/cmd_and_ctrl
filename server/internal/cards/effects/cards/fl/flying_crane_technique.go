package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Flying Crane Technique — Instant {3}{U}{R}{W} (Khans of Tarkir):
//
//	"Untap all creatures you control. They gain flying and double strike
//	 until end of turn."
//
// Terror of Mount Velus's grant over a Match: the creatures you control are
// locked in as it resolves (CR 611.2c), so a creature that enters later this
// turn gets neither keyword. The untap comes first and the same set gains the
// keywords.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "44cf99b0-8a5e-4df9-abc5-552f8b838c56",
		Name:         "Flying Crane Technique",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerCombatGrant},
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error { return effects.UntapAllCreaturesYouControl{}.Apply(ctx) }),
			effects.Instruction(func(ctx *effects.Context) error {
				return effects.GrantKeywordUntilEOT{
					Match:    effects.And(effects.Creature(), effects.YouControl()),
					Keywords: []string{"flying", "double strike"},
					Label:    "Flying Crane Technique — flying and double strike",
				}.Apply(ctx)
			}),
		),
	})
}
