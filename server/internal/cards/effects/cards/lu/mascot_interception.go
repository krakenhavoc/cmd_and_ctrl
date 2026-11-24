package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Mascot Interception — Sorcery {3}{R} (Strixhaven: School of Mages, #5502):
//
//	"This spell costs {3} less to cast if it targets a creature token.
//	 Gain control of target creature until end of turn. Untap that creature.
//	 It gets +2/+0 and gains haste until end of turn."
//
// Bloody Betrayal's threaten with Price of Loyalty's +2/+0, and Mystical
// Dispute's discount judged on the announced target as the cost is determined
// (CR 601.2f, after targets are chosen in 601.2c): a creature token target
// makes it {R}. The control, the untap, the pump and haste last until end of
// turn; a target that has become illegal makes the spell do nothing
// (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ae3da0be-2a3f-498a-8854-73d0a5526ec2",
		Name:         "Mascot Interception",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		SelfCostModifiers: []game.CostModifier{
			effects.CostsLessIfItTargets(3, "This spell costs {3} less to cast if it targets a creature token.",
				effects.And(effects.Creature(), effects.IsTokenPredicate())),
		},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GainControl{
				Target:   target,
				Duration: effects.DurationUntilEndOfTurn(ctx),
				Label:    "Mascot Interception — gain control until end of turn",
			}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: target}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.BoostUntilEOT{Target: target, Power: 2, Label: "Mascot Interception — +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{Target: target, Keywords: []string{"haste"},
				Label: "Mascot Interception — haste until end of turn"}.Apply(ctx)
		},
	})
}
