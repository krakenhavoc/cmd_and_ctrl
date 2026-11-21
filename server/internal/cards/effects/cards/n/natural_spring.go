package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Natural Spring — Sorcery {3}{G}{G} (Portal, #5630):
//
//	"Target player gains 8 life."
//
// Sylvan Bounty's spell half: the shared targetPlayerGainsLife body
// (target_player_gains_life.go) with 8 life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f7571a2e-aaf3-4148-ab76-2a2e35273c70",
		Name:         "Natural Spring",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 8}),
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.TargetPlayerGainsLife(8),
	})
}
