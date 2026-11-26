package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Slave of Bolas — Sorcery {3}{U/R}{B} (Alara Reborn, #5586):
//
//	"Gain control of target creature. Untap that creature. It gains haste until
//	 end of turn. Sacrifice it at the beginning of the next end step."
//
// Angrath's threaten shape with an unlimited control change: control, untap and
// haste in that order, then a delayed trigger at the next end step sacrifices
// the same object (CR 400.7: a creature that left and came back is not it).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "07190a8b-2a59-45b3-9fbf-505161698a9c",
		Name:         "Slave of Bolas",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GainControl{Target: id, Duration: game.IndefiniteDuration(), Label: "Slave of Bolas — gain control"}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.UntapTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"haste"}, Label: "Slave of Bolas — haste"}).Apply(ctx); err != nil {
				return err
			}
			c, ok := effects.BattlefieldCardForEffect(ctx.Game, id)
			if !ok {
				return nil
			}
			return effects.ScheduleDelayedTrigger{
				At:         game.StepEnd,
				Controller: ctx.Controller(),
				Label:      "Slave of Bolas — sacrifice it",
				Body:       slaveOfBolasSacrificeBody,
				Params:     game.EffectParams{Object: game.ObjectRef{ID: id, Epoch: c.ObjectEpoch}},
			}.Apply(ctx)
		},
	})
}

// slaveOfBolasSacrificeBody is the delayed sacrifice. An on-disk key: never
// renamed, never reused.
var slaveOfBolasSacrificeBody = game.DelayedBody("slave-of-bolas/sacrifice", func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
	c, ok := g.LookupCardForEffect(p.Object.ID)
	if !ok || !effects.OnBattlefield(g, p.Object.ID) || c.ObjectEpoch != p.Object.Epoch {
		return nil
	}
	return effects.SacrificePermanent{Target: p.Object.ID}.Apply(effects.NewContext(g, item))
})
