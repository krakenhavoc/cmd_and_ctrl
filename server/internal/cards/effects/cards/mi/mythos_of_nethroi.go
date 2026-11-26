package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Mythos of Nethroi — Instant {2}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Destroy target nonland permanent if it's a creature or if {G}{W} was
//	 spent to cast this spell."
//
// River's Grasp's mana-spent reading over the spell's own payment
// (ReadsManaSpent: a waived payment asks what was paid, a copy has no mana
// spent): the {G}{W} condition holds when at least one green and one white
// mana were spent. The creature test is read from the target as it resolves
// (CR 608.2b). WantsColors names both colours so an auto-paid cast reaches
// for them.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:       "16da72a3-d980-4dd8-99f2-8191cce00978",
		Name:           "Mythos of Nethroi",
		Completeness:   effects.CompletenessFull,
		Purpose:        game.Purpose{Answers: game.AnswerRemove},
		ReadsManaSpent: game.ReadsOwnPayment,
		WantsColors:    []string{"G", "W"},
		Targets:        effects.TargetPermanent("target nonland permanent", effects.Nonland()),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			t, ok := ctx.ClauseTarget(0)
			if !ok || !ctx.IsTargetLegal(t) {
				return nil
			}
			card, ok := ctx.Game.LookupCardForEffect(t.ID)
			if !ok {
				return nil
			}
			if !card.IsCreature() && (ctx.ManaSpentOfColor("G") == 0 || ctx.ManaSpentOfColor("W") == 0) {
				return nil
			}
			return effects.DestroyTheTargetPermanent(item, ctx)
		},
	})
}
