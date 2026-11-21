package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Inferno Jet — Sorcery {5}{R} (Hour of Devastation, #5534):
//
//	"Inferno Jet deals 6 damage to target opponent or planeswalker.
//	 Cycling {2} ({2}, Discard this card: Draw a card.)"
//
// The target is an opponent (never you) or a planeswalker
// (targetOpponentOrPlaneswalker, Zealot of the God-Pharaoh's), and the
// shared Cycling row.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "50c13b38-9865-4c60-b2df-3fd6e3a9c6a2",
		Name:         "Inferno Jet",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetOpponentOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 6)), Answers: game.AnswerRemove},
		OnResolve:    effects.DamageToFirstTargetStep(6),
		Activated:    []effects.ActivatedAbility{effects.Cycling("{2}")},
	})
}
