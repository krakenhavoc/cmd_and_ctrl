package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hoodwink — Instant {1}{U}:
//
//	"Return target artifact, enchantment, or land to its owner's hand."
//
// The shared bounceTheTarget body: the permanent goes to its OWNER's hand,
// and a target gone by resolution does nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a2fe10b8-857e-4881-85ef-24a0f96a3d79",
		Name:         "Hoodwink",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target artifact, enchantment, or land", effects.Or(effects.Artifact(), effects.Enchantment(), effects.Land())),
		OnResolve:    effects.BounceTheTarget,
	})
}
