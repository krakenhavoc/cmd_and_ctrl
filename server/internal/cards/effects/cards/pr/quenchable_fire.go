package cardspr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Quenchable Fire — Sorcery {3}{R} (Conflux):
//
//	"Quenchable Fire deals 3 damage to target player or planeswalker. It deals
//	 an additional 3 damage to that player or planeswalker at the beginning of
//	 your next upkeep step unless that player or that planeswalker's controller
//	 pays {U} before that step."
//
// The first 3 damage is dealt as the spell resolves. The additional damage is a
// delayed trigger at the beginning of the caster's next upkeep (CR 603.7), the
// delayed trigger with the payment on it (#6220): the player, or the
// planeswalker's controller, may pay {U} at any time they have priority until
// that step begins, and paying removes the trigger. The target is recorded as
// the object it was (CR 400.7): a planeswalker that has left, or come back as a
// new object, takes nothing. The payer named is the planeswalker's controller
// as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d2eeb1eb-38c0-4b76-ad96-a53b6bdc00e1",
		Name:         "Quenchable Fire",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 3)),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok {
				return nil
			}
			if err := (effects.DealDamage{Source: ctx.Source(), Target: t.ID, Amount: 3}).Apply(ctx); err != nil {
				return err
			}
			params := game.EffectParams{}
			payer := t.ID
			if t.Kind == game.TargetPlayer {
				params.Player = t.ID
			} else {
				ref := game.ObjectRef{ID: t.ID, Epoch: ctx.Game.ObjectEpochForEffect(t.ID)}
				params.Object = ref
				info, found := ctx.Game.PermanentForEffect(ref)
				if !found {
					return nil
				}
				payer = info.Controller
			}
			return effects.ScheduleDelayedTrigger{
				Label:       "Quenchable Fire — 3 more damage unless they pay {U}",
				At:          game.StepUpkeep,
				TurnOf:      ctx.Controller(),
				Body:        quenchableFireDamage,
				Params:      params,
				PayToCancel: &game.PayToCancel{Player: payer, Cost: "{U}"},
			}.Apply(ctx)
		},
	})
}

// quenchableFireDamage is the delayed trigger: 3 more damage to the player, or
// to the planeswalker while it is still the object that was targeted. An
// on-disk key: never renamed, never reused.
var quenchableFireDamage = game.DelayedBody("old-b507/quenchable-fire-damage",
	func(g *game.Game, item *game.StackItem, p game.EffectParams) error {
		ctx := effects.NewContext(g, item)
		target := p.Player
		if target == uuid.Nil {
			if p.Object.ID == uuid.Nil || !effects.OnBattlefield(g, p.Object.ID) || g.ObjectEpochForEffect(p.Object.ID) != p.Object.Epoch {
				return nil
			}
			target = p.Object.ID
		}
		return effects.DealDamage{Source: item.SourceCardID, Target: target, Amount: 3}.Apply(ctx)
	})
