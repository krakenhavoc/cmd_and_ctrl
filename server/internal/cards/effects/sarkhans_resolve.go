package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "30f058f8-5777-4c25-b881-8f1d0ce17d3a",
		Name:         "Sarkhan's Resolve",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(Mode("Target creature gets +3/+3 until end of turn.",
				TargetCreature("target creature")), ForTargets(BenefitUntilEOT(0, 3, 3))),
			Mode("Destroy target creature with flying.",
				TargetCreature("target creature with flying", HasKeyword("flying"))),
		),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := ModeTarget(ctx, 0)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			if ctx.HasMode(1) {
				return DestroyTarget{Target: t.ID}.Apply(ctx)
			}
			return BoostUntilEOT{Target: t.ID, Power: 3, Toughness: 3, Label: "Sarkhan's Resolve — +3/+3"}.Apply(ctx)
		},
	})
}
