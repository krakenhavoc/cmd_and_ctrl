package cardsx

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Yip Yip! — Instant — Lesson {W} (Avatar: The Last Airbender, #4459):
//
//	"Target creature you control gets +2/+2 until end of turn. If that
//	 creature is an Ally, it also gains flying until end of turn."
//
// "Is an Ally" is read as the spell resolves, after the pump, through the
// layered view, so a creature made an Ally by an effect counts. "Ally" is
// a word of the spell's text (TextSubtype, CR 612.2). Lesson is a spell
// subtype with no rules of its own.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5453550f-9136-4568-baf8-c3d20e33fbb6",
		Name:         "Yip Yip!",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 2)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target := effects.FirstLegalBattlefieldTarget(ctx)
			if target == uuid.Nil {
				return nil
			}
			if err := (effects.BoostUntilEOT{Target: target, Power: 2, Toughness: 2, Label: "Yip Yip! — +2/+2"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := ctx.Game.LookupCardForEffect(target)
			if !ok || !c.HasSubtype(ctx.TextSubtype("Ally")) {
				return nil
			}
			return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"flying"}, Label: "Yip Yip! — flying"}.Apply(ctx)
		},
	})
}
