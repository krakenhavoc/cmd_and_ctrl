package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Duty Beyond Death — Instant {1}{W} (Secrets of Strixhaven Commander,
// #4834):
//
//	"As an additional cost to cast this spell, sacrifice a creature.
//	 Creatures you control gain indestructible until end of turn. Put a
//	 +1/+1 counter on each creature you control. (Damage and effects
//	 that say "destroy" don't destroy those creatures.)"
//
// The sacrifice is paid while the spell is cast (CR 601.2h), so the
// creature given up gets neither the keyword nor a counter. Both
// sentences work on the creatures you control as the spell resolves
// (CR 611.2c for the grant): the set is fixed then, so a creature that
// arrives later this turn does not gain indestructible. The counters
// wait for the prompt the grant may ask, and each creature gets one
// (Unbreakable Formation's pair of sentences).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:       "67e6ce99-4883-4713-9822-cb2334765b5b",
		Name:           "Duty Beyond Death",
		Purpose:        game.Purpose{Answers: game.AnswerSurvive},
		Completeness:   CompletenessFull,
		AdditionalCost: SacrificeCost("a creature", Creature()),
		OnResolve: ResolveSteps(
			GrantKeywordUntilEOT{
				Match:    And(Creature(), YouControl()),
				Keywords: []string{"indestructible"},
				Label:    "Duty Beyond Death — indestructible",
			},
			Instruction(func(ctx *Context) error {
				return b11PutCounterOnEachCreatureYouControl(ctx.Game, ctx.Item)
			}),
		),
	})
}
