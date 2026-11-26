package cardsco

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Concussive Bolt — Sorcery {3}{R}{R} (Mirrodin Besieged, S82 OLD b465, #5576):
//
//	"Concussive Bolt deals 4 damage to target player or planeswalker.
//	 Metalcraft — If you control three or more artifacts, creatures controlled
//	 by that player or by that planeswalker's controller can't block this
//	 turn."
//
// Two instructions in printed order. The player whose creatures are barred is
// read BEFORE the damage, since the planeswalker may die to it: the target
// player, or the planeswalker's controller. Metalcraft is read as the spell
// resolves. "Can't block" is a rules effect (CR 611.2c does not lock its set):
// the scope reads that player's creatures live until cleanup (Headliner
// Scarlett's shape), so one they get later can't block either.
//
// No simplification.
func init() {
	const label = "Concussive Bolt — creatures that player controls can't block this turn"
	effects.Register(effects.Spec{
		OracleID:     "e99958bf-458c-403b-aae2-91d60c055ee2",
		Name:         "Concussive Bolt",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      effects.ForTargets(effects.DamageToTarget(0, 4)),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := effects.FirstLegalTarget(ctx)
			if !ok {
				return nil
			}
			barred := t.ID
			if t.Kind == game.TargetCard {
				c, found := ctx.Game.LookupCardForEffect(t.ID)
				if !found {
					return nil
				}
				barred = c.Controller
			}
			if err := effects.DamageToFirstTargetStep(4)(item, ctx); err != nil {
				return err
			}
			if barred == uuid.Nil || !effects.ControlsAtLeast(3, effects.MatchArtifact)(ctx.Game, ctx.Controller(), ctx.Source()) {
				return nil
			}
			ctx.Game.RegisterScopedRuleEffectForEffect(ctx.Source(), game.ScopeYourCreatures, barred,
				[]game.Mod{game.AddRestrictionsMod(game.CantBlock)}, effects.DurationUntilEndOfTurn(ctx), label)
			return nil
		},
	})
}
