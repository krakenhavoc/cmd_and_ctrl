package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Stolen by the Fae — Sorcery {X}{U}{U} (Bloomburrow Commander, #5368):
//
//	"Return target creature with mana value X to its owner's hand. You
//	 create X 1/1 blue Faerie creature tokens with flying."
//
// The target clause is Entrancing Melody's: the creature's mana value
// must equal the X the spell was cast for, checked as it is chosen and
// again on resolution (CR 608.2b). With the target illegal the spell does
// nothing at all, tokens included. The tokens are made after the return
// has settled (the bounce's tail), so a commander's command-zone question
// is answered first; the clauses are independent, so they are made even
// if the creature did not reach a hand. X is zero for a mana value of 0
// (a token creature): no tokens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5bafd0e0-8bf8-4e06-a54a-81cc6cc4bdb7",
		Name:         "Stolen by the Fae",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		Targets:      effects.TargetCreature("target creature with mana value X").WithManaValueEqualsX(),
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return effects.BounceToHand{Target: id, ThenTail: ctx.ItemTail(effects.Cmdr2B11StolenByTheFaeTokens)}.Apply(ctx)
		},
	})
}
