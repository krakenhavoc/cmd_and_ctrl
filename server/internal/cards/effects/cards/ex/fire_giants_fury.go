package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Fire Giant's Fury — Sorcery {1}{R} (Kaldheim):
//
//	"Target Giant you control gets +2/+2 and gains trample until end of turn.
//	 Whenever it deals combat damage to a player this turn, exile that many
//	 cards from the top of your library. Until the end of your next turn, you
//	 may play those cards."
//
// The pump is until end of turn; the rider is a repeating delayed trigger (CR
// 603.7b) bound to that Giant's object (it names the creature, not the
// spell), one firing per combat damage event it deals to a player, exiling
// that many cards with a window to play them until the end of the
// controller's next turn. The Giant gone by resolution fizzles the spell.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e5f83bfc-ff74-4bb9-82c8-a07fedc6e916",
		Name:         "Fire Giant's Fury",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.BenefitUntilEOT(0, 2, 2, "trample"))},
		Targets:      effects.TargetCreature("target Giant you control", effects.HasSubtype("Giant"), effects.YouControl()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			if err := effects.PumpAndGrantFirstTarget(2, 2, "Fire Giant's Fury — +2/+2 and trample", "trample")(ctx.Game, item); err != nil {
				return err
			}
			ids := effects.LegalTargetCardIDs(ctx)
			if len(ids) == 0 {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(ids[0])
			if !ok {
				return nil
			}
			return effects.DelayedOnEvent{
				Label:      "Fire Giant's Fury — exile that many cards from the top of your library; until the end of your next turn, you may play those cards",
				On:         []game.EventKind{game.EventDealDamage},
				Condition:  effects.ThatCreatureHitAPlayerCondition,
				CondParams: game.EffectParams{Object: game.ObjectRef{ID: ids[0], Epoch: c.ObjectEpoch}},
				Body:       fireGiantsFuryExile,
				Repeats:    true,
			}.Apply(ctx)
		},
	})
}

// fireGiantsFuryExile is the delayed trigger's body: that many cards (the
// combat damage the Giant dealt) from the top of the controller's library,
// each playable until the end of their next turn. An on-disk key: never
// renamed, never reused.
var fireGiantsFuryExile = game.SimpleDelayedBody("exile/s82-old-b061-fire-giants-fury", func(g *game.Game, item *game.StackItem) error {
	if item.Trigger == nil || item.Trigger.Event.Amount <= 0 {
		return nil
	}
	_, err := g.ExileTopWithPermissionForEffect(item.Controller, item.Controller, item.Trigger.Event.Amount, game.CastPermission{
		Duration: g.UntilEndOfYourNextTurnDuration(item.Controller),
	})
	return err
})
