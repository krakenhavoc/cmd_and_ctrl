package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Boon of Erebos — Instant {B} (Theros, S82 OLD b391, #5561):
//
//	"Target creature gets +2/+0 until end of turn. Regenerate it. You lose
//	 2 life."
//
// Three steps in printed order, the shield on the same target (CR 701.19a);
// the life loss is the caster's own and happens whether or not the creature
// is still there to be regenerated (a target that has become illegal counters
// the whole spell, CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "55821968-baf2-4fa6-8376-9ec305e27b3b",
		Name:         "Boon of Erebos",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 2, 0)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if t, ok := ctx.ClauseTarget(0); ok && t.Kind == game.TargetCard {
				if err := (effects.BoostUntilEOT{Target: t.ID, Power: 2, Label: "Boon of Erebos — +2/+0"}).Apply(ctx); err != nil {
					return err
				}
				if err := (effects.Regenerate{Target: t.ID}).Apply(ctx); err != nil {
					return err
				}
			}
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		},
	})
}
