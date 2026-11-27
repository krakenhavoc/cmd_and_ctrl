package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Hypnotic Cloud (Invasion, S82 OLD batch 648, #5616):
//
//	Sorcery {1}{B}:
//	"Kicker {4}
//	Target player discards a card. If this spell was kicked, that player discards three cards instead."
func init() {
	effects.Register(effects.Spec{
		OracleID:      "277757c6-4c4c-4a4c-84f6-4a64fe3b3bfe",
		Name:          "Hypnotic Cloud",
		Completeness:  effects.CompletenessFull,
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{4}")},
		Targets:       effects.TargetPlayer("target player"),
		OnResolve:     effects.S82OldB648HypnoticCloudResolve,
	})
}
