package cardsco

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Compelling Argument — {1}{U} Sorcery (Amonkhet, S82 old batch 245):
//
//	"Target player mills five cards.
//	 Cycling {U}"
//
// The target player is read as the spell resolves; a player who left the game
// is an illegal target and the spell does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "686e0b86-121c-45be-842d-c83d3b6547f6",
		Name:         "Compelling Argument",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayer("target player"),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{U}")},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			who, ok := effects.FirstLegalPlayerTarget(ctx)
			if !ok {
				return nil
			}
			return effects.MillCards{Player: who, N: 5}.Apply(ctx)
		},
	})
}
