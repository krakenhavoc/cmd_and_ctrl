package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Arrester's Zeal — Instant {W} (Ravnica Remastered):
//
//	"Target creature gets +2/+2 until end of turn.
//	 Addendum — If you cast this spell during your main phase, that creature
//	 gains flying until end of turn."
//
// Addendum is read as the spell resolves: it is the caster's own main phase
// (Code of Constraint's b14InYourMainPhase), which is also when it was cast,
// as a spell cannot change phase while on the stack.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6b70af46-ff27-4eec-9692-248b06248de6",
		Name:         "Arrester's Zeal",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: effects.ResolveSteps(
			effects.Instruction(func(ctx *effects.Context) error {
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.BoostUntilEOT{Target: id, Power: 2, Toughness: 2, Label: "Arrester's Zeal — +2/+2"}.Apply(ctx)
			}),
			effects.Instruction(func(ctx *effects.Context) error {
				if !effects.B14InYourMainPhase(ctx.Game, ctx.Controller()) {
					return nil
				}
				id, ok := effects.B16FirstLegalTargetCard(ctx)
				if !ok {
					return nil
				}
				return effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"flying"},
					Label: "Arrester's Zeal — addendum: flying until end of turn"}.Apply(ctx)
			}),
		),
	})
}
