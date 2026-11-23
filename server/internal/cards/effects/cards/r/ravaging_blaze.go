package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ravaging Blaze — Instant {X}{R}{R} (Commander Masters, #5948):
//
//	"Ravaging Blaze deals X damage to target creature.
//	 Spell mastery — If there are two or more instant and/or sorcery cards in
//	 your graveyard, Ravaging Blaze also deals X damage to that creature's
//	 controller."
//
// Spell mastery is read as the spell resolves, so Ravaging Blaze itself is
// not yet in the graveyard (CR 608.2n). The controller is the creature's
// controller as the spell resolves, or as it last existed if it has gone
// (CR 608.2h); a creature that left in response makes the whole spell fizzle
// instead (CR 608.2b), since its only target is illegal.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a41760a2-d359-4581-b8d3-c1b572ac15c2",
		Name:         "Ravaging Blaze",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetCreature("target creature"),
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			target, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(target)
			if !ok {
				return nil
			}
			controller := card.Controller
			x := ctx.X()
			if err := (effects.DealDamage{Source: ctx.Source(), Target: target, Amount: x}).Apply(ctx); err != nil {
				return err
			}
			if !effects.SpellMastery(ctx.Game, ctx.Controller()) {
				return nil
			}
			return effects.DealDamage{Source: ctx.Source(), Target: controller, Amount: x}.Apply(ctx)
		},
	})
}
