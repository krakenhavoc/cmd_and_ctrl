package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sylvan Bounty — Instant {5}{G} (Conflux, #5587):
//
//	"Target player gains 8 life.
//	 Basic landcycling {1}{G} ({1}{G}, Discard this card: Search your
//	 library for a basic land card, reveal it, put it into your hand,
//	 then shuffle.)"
//
// Heroes' Reunion at 8 life, to any player, with the engine's basic
// landcycling row (CR 702.29e). A target player who has left the game
// counters the spell (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "c20635d2-20e9-48f7-88f0-6de0f953bb08",
		Name:         "Sylvan Bounty",
		Completeness: effects.CompletenessFull,
		Purpose:      effects.ForTargets(game.TargetPurpose{Slot: 0, LifeGain: 8}),
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.TargetPlayerGainsLife(8),
		Activated:    []effects.ActivatedAbility{effects.BasicLandcycling("{1}{G}")},
	})
}
