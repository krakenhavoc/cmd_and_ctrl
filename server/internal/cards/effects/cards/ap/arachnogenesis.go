package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arachnogenesis — Instant {2}{G}:
//
//	"Create X 1/2 green Spider creature tokens with reach, where X is the
//	 number of creatures attacking you. Prevent all combat damage that
//	 would be dealt this turn by non-Spider creatures."
//
// X counts the creatures attacking you, not your planeswalkers, as the
// spell resolves (CR 608.2h). The shield is #2026's negation over a
// subtype, read as each creature would deal combat damage (CR 609.7b):
// the new Spiders deal theirs, and a changeling is a Spider (CR
// 702.73a).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b655bee5-52d3-467e-b16c-cfc2edf2b1a1",
		Name:         "Arachnogenesis",
		Completeness: effects.CompletenessFull,
		OnResolve: effects.ResolveSteps(
			// The shield waits for a prompt the Spiders' creation asks (CR 608.2c, #4155).
			effects.Instruction(func(ctx *effects.Context) error {
				n := effects.CreaturesAttackingPlayer(ctx.Game, ctx.Controller())
				if n <= 0 {
					return nil
				}
				return effects.CreateToken{Template: effects.TokenCard("1/2 green Spider with reach"), N: n}.Apply(ctx)
			}),
			effects.CombatShieldAgainstCreatures(effects.ExceptSubtypes("Spider")),
		),
	})
}
