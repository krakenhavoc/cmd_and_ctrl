package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Extortion — Sorcery {3}{B}{B} (Mercadian Masques, S82 OLD batch 697, #5619):
//
//	"Look at target player's hand and choose up to two cards from it. That
//	 player discards those cards."
//
// ChooseFromRevealedHand with Look: only the caster sees the hand, and picks
// from none to two cards.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8841fdb9-0565-4ef5-814c-293c14242741",
		Name:         "Extortion",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ChooseFromRevealedHand{Player: effects.TargetedPlayer(ctx), Label: "card", Count: 2, UpTo: true, Look: true}.Apply(ctx)
		},
	})
}
