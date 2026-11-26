package cardsri

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ruthless Instincts — {2}{G} Instant (Fate Reforged):
//
//	"Choose one —
//	 • Target nonattacking creature gains reach and deathtouch until end of
//	   turn. Untap it.
//	 • Target attacking creature gets +2/+2 and gains trample until end of
//	   turn."
//
// Each bullet has its own target, chosen as the spell is cast (CR 700.2) and
// re-checked as it resolves (CR 608.2b): a creature that is attacking when the
// first bullet resolves is no longer a legal target for it.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b60aff46-2d5a-47e2-9340-39d34a47c3e1",
		Name:         "Ruthless Instincts",
		Completeness: effects.CompletenessFull,
		Modes: effects.ChooseOne(
			effects.ModeWithPurpose(effects.ModeDoing("Target nonattacking creature gains reach and deathtouch until end of turn. Untap it.",
				effects.TargetCreature("target nonattacking creature", effects.Not(effects.AttackingCreature())),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"reach", "deathtouch"}, Label: "Ruthless Instincts — reach and deathtouch"}).Apply(ctx); err != nil {
						return err
					}
					return effects.UntapTarget{Target: t.ID}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "reach", "deathtouch"))),
			effects.ModeWithPurpose(effects.ModeDoing("Target attacking creature gets +2/+2 and gains trample until end of turn.",
				effects.TargetCreature("target attacking creature", effects.AttackingCreature()),
				func(_ *game.StackItem, ctx *effects.Context, occ int) error {
					t, ok := effects.ModeTarget(ctx, occ)
					if !ok {
						return nil
					}
					if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Toughness: 2, Label: "Ruthless Instincts — +2/+2"}).Apply(ctx); err != nil {
						return err
					}
					return effects.GrantKeywordUntilEOT{Target: t.ID, Keywords: []string{"trample"}, Label: "Ruthless Instincts — trample"}.Apply(ctx)
				}), effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample"))),
		),
	})
}
