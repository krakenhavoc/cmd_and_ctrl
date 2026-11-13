package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "64de050c-7167-4ebd-8584-f064d8be3302",
		Name:         "Armor of Shadows",
		Purpose:      ForTargets(BenefitUntilEOT(0, 1, 0, "indestructible")),
		Completeness: CompletenessFull,
		Targets:      TargetCreature("target creature"),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return untilEndOfTurn(ctx, id, nil, "Armor of Shadows — +1/+0 and indestructible",
				game.ModifyPTMod(1, 0), game.AddKeywordsMod("indestructible"))
		},
	})
}
