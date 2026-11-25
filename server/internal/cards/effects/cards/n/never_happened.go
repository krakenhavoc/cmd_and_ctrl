package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Never Happened — Sorcery {2}{B} (Guilds of Ravnica, #5525):
//
//	"Target opponent reveals their hand. You choose a nonland card from that
//	 player's graveyard or hand and exile it."
//
// Memory Leak's body without the cycling: the hand is revealed and the card
// is chosen from the revealed hand and the graveyard together; a land card is
// never a choice from either. It is exiled, not discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ec4a1a8d-4071-40e9-812e-42c746180d99",
		Name:         "Never Happened",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileNonlandFromHandOrGraveyard(ctx)
		},
	})
}
