package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Memory Leak — Sorcery {2}{B} (Ikoria: Lair of Behemoths, #5512):
//
//	"Target opponent reveals their hand. You choose a nonland card from that
//	 player's graveyard or hand and exile it.
//	 Cycling {1}"
//
// The hand is revealed and the card is chosen from the revealed hand and the
// graveyard together (Agonizing Remorse's shape, with no life loss); a land
// card is never a choice from either. It is exiled, not discarded.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "256d884f-5dea-4906-a53d-cf14d43e6bc8",
		Name:         "Memory Leak",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target opponent", effects.Opponent()),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{1}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.ExileNonlandFromHandOrGraveyard(ctx)
		},
	})
}
