package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vigorous Charge — Instant {G} (Invasion, S82 OLD #5616):
//
//	"Kicker {W} (You may pay an additional {W} as you cast this spell.)
//	 Target creature gains trample until end of turn. Whenever that creature
//	 deals combat damage this turn, if this spell was kicked, you gain life
//	 equal to that damage."
//
// The trample is a plain until-end-of-turn grant. The life clause exists only
// when the spell was kicked (CR 603.4, read as it resolves): it is a delayed
// trigger on that object (a creature that left is a new object, CR 400.7)
// that fires on each combat damage event it deals this turn (CR 603.7c) and
// ends with the turn (CR 514.2). Each event gains its own damage, to a
// player, planeswalker, battle or creature alike, for the spell's
// controller.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b8ce3f84-2053-4d16-ab3d-e552bd8d0e8c",
		Name:          "Vigorous Charge",
		Completeness:  effects.CompletenessFull,
		Purpose:       effects.ForTargets(effects.BenefitUntilEOT(0, 0, 0, "trample")),
		Targets:       effects.TargetCreature("target creature"),
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{W}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			if err := (effects.GrantKeywordUntilEOT{Target: id, Keywords: []string{"trample"}, Label: "Vigorous Charge — trample"}).Apply(ctx); err != nil {
				return err
			}
			if !ctx.WasKicked() {
				return nil
			}
			ref, ok := ctx.Game.PermanentRefForEffect(id)
			if !ok {
				return nil
			}
			return effects.DelayedOnEvent{
				Label:      "Vigorous Charge — gain life equal to that damage",
				On:         []game.EventKind{game.EventDealDamage},
				Condition:  vigorousChargeDealt,
				CondParams: game.EffectParams{Object: ref},
				Body:       vigorousChargeGain,
				Repeats:    true,
			}.Apply(ctx)
		},
	})
}

var vigorousChargeDealt = game.DelayedCondition("combat/vigorous-charge-creature-dealt-combat-damage", func(ev game.Event, _ *game.DelayedTrigger, g *game.Game, p game.EffectParams) bool {
	if ev.Kind != game.EventDealDamage || !ev.Combat || ev.Amount <= 0 || ev.Source != p.Object.ID {
		return false
	}
	ref, ok := g.PermanentRefForEffect(ev.Source)
	return ok && ref.Epoch == p.Object.Epoch
})

var vigorousChargeGain = game.SimpleDelayedBody("vigorous-charge/gain-that-much", func(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil {
		return nil
	}
	return effects.GainLife{Player: item.Controller, Amount: item.Trigger.Event.Amount}.Apply(effects.NewContext(g, item))
})
