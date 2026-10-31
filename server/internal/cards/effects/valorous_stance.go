package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Valorous Stance — Instant {1}{W} (Marvel Super Heroes Commander,
// #4455):
//
//	"Choose one —
//	 • Target creature gains indestructible until end of turn.
//	 • Destroy target creature with toughness 4 or greater."
//
// Each bullet carries its own target clause, so choosing one asks for
// that clause only; the destroy bullet is the shared DestroyTheModesTarget.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1f938353-9081-4dc1-b8e7-d18ece038bd4",
		Name:         "Valorous Stance",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeWithPurpose(ModeDoing("Target creature gains indestructible until end of turn.",
				TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *Context, occ int) error {
					t, ok := ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Valorous Stance — indestructible"}.Apply(ctx)
				}), ForTargets(BenefitUntilEOT(0, 0, 0, "indestructible"))),
			ModeDoing("Destroy target creature with toughness 4 or greater.",
				TargetCreature("target creature with toughness 4 or greater", ToughnessGE(4)),
				DestroyTheModesTarget),
		),
	})
}
