package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

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
	effects.Register(effects.Spec{
		OracleID:     "1f938353-9081-4dc1-b8e7-d18ece038bd4",
		Name:         "Valorous Stance",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target creature gains indestructible until end of turn.",
				effects.TargetCreature("target creature"),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"indestructible"}, Label: "Valorous Stance — indestructible"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "indestructible"))),
			effects.ModeDoing("Destroy target creature with toughness 4 or greater.",
				effects.TargetCreature("target creature with toughness 4 or greater", effects.ToughnessGE(4)),
				effects.DestroyTheModesTarget),
		),
	})
}
