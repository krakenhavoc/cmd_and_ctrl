package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Armor of Shadows — Instant {B} (Commander Legends: Battle for Baldur's
// Gate, #5490):
//
//	"Until end of turn, target creature gets +1/+0 and gains
//	 indestructible. (Damage and effects that say "destroy" don't destroy
//	 it.)"
//
// One sentence, so one until-end-of-turn record carries both halves:
// the +1/+0 in layer 7c and indestructible in layer 6 (Moment of Valor's
// shape, without the untap). Any creature may be the target, an
// opponent's included. A target that has left or become illegal by
// resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "64de050c-7167-4ebd-8584-f064d8be3302",
		Name:         "Armor of Shadows",
		Purpose:      effects.ForTargets(effects.BenefitUntilEOT(0, 1, 0, "indestructible")),
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.UntilEndOfTurn(ctx, id, nil, "Armor of Shadows — +1/+0 and indestructible",
				game.ModifyPTMod(1, 0), game.AddKeywordsMod("indestructible"))
		},
	})
}
