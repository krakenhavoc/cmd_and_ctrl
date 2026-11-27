package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Intimidation Bolt — Instant {1}{R}{W} (Alara Reborn):
//
//	"Intimidation Bolt deals 3 damage to target creature. Other creatures
//	 can't attack this turn."
//
// "Other" is every creature but the target, read live (CR 611.2c: a rule
// change), so one that enters later this turn can't attack either. The
// target is named as the object it is before the damage (CR 400.7): a
// target the bolt kills is no longer a creature on the battlefield and the
// rule then covers every creature; a card that returns to the battlefield
// is a new object and is covered. Reuses Last Night Together's
// ScopeCreaturesExceptObjects record.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3018c0ef-876a-4254-9513-42f351e1da3e",
		Name:         "Intimidation Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			id := item.Targets[0].ID
			var refs []game.ObjectRef
			if c, ok := ctx.Game.LookupCardForEffect(id); ok {
				refs = append(refs, game.ObjectRef{ID: id, Epoch: c.ObjectEpoch})
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: id, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			if len(refs) == 0 {
				return nil
			}
			ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeCreaturesExceptObjects, ctx.Controller(),
				[]game.Mod{game.OnlyTheseCanAttackMod(refs)}, effects.DurationUntilEndOfTurn(ctx),
				"Intimidation Bolt — other creatures can't attack this turn")
			return nil
		},
	})
}
