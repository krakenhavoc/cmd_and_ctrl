package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sarkhan's Resolve — Instant {1}{G} (Tarkir: Dragonstorm, #4463):
//
//	"Choose one —
//	 • Target creature gets +3/+3 until end of turn.
//	 • Destroy target creature with flying."
//
// Two bullets, each with its own target, chosen as the spell is cast
// (CR 700.2a). A target that is no longer legal on resolution (for the
// second bullet, one that has lost flying) is left alone (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "30f058f8-5777-4c25-b881-8f1d0ce17d3a",
		Name:         "Sarkhan's Resolve",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.Mode("Target creature gets +3/+3 until end of turn.",
				effects.TargetCreature("target creature")), effects.ForTargets(effects.BenefitUntilEOT(0, 3, 3))),
			effects.Mode("Destroy target creature with flying.",
				effects.TargetCreature("target creature with flying", effects.HasKeyword("flying"))),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.ModeTarget(ctx, 0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if ctx.HasMode(1) {
				return effects.DestroyTarget{Target: t.ID}.Apply(ctx)
			}
			return effects.BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Sarkhan's Resolve — +3/+3"}.Apply(ctx)
		},
	})
}
