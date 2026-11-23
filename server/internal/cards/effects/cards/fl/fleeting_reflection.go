package cardsfl

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fleeting Reflection — Instant {1}{U} (Outlaws of Thunder Junction):
//
//	"Target creature you control gains hexproof until end of turn. Untap that
//	 creature. Until end of turn, it becomes a copy of up to one other target
//	 creature."
//
// Two clauses: your creature, then up to one OTHER creature (Distinct, any
// controller). In printed order: the hexproof, the untap, then the copy
// (Hulkling's BecomeCopy until end of turn, no exceptions: the creature
// becomes the other creature exactly). With no second target it still gains
// hexproof and untaps. Either target gone by resolution is skipped (CR 608.2b);
// the copy takes the other creature's values as it resolves (CR 707.2).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c915f815-77d7-4e08-b178-82b29a58bb03",
		Name:         "Fleeting Reflection",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "hexproof")),
		Targets: effects.Clauses(
			effects.TargetCreature("target creature you control", effects.YouControl()),
			effects.Distinct(effects.TargetCreature("up to one other target creature").WithCount(0, 1)),
		),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			mine, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(mine) {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: mine.ID, Keywords: []string{"hexproof"}, Label: "Fleeting Reflection — hexproof"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: mine.ID}).Apply(ctx); err != nil {
				return err
			}
			var other uuid.UUID
			for _, t := range ctx.ClauseTargets(1) {
				if ctx.IsTargetLegal(t) {
					other = t.ID
					break
				}
			}
			if other == uuid.Nil {
				return nil
			}
			return effects.BecomeCopy{
				Targets: []uuid.UUID{mine.ID},
				Of:      other,
				Label:   "Fleeting Reflection — becomes a copy until end of turn",
			}.Apply(ctx)
		},
	})
}
