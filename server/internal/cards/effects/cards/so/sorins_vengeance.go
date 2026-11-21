package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sorin's Vengeance — Sorcery {4}{B}{B}{B} (Magic 2012, #5573):
//
//	"Sorin's Vengeance deals 10 damage to target player or
//	 planeswalker and you gain 10 life."
//
// Vicious Hunger's body at 10 against Lava Axe's "target player or
// planeswalker" clause.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "75d9c036-4f0d-4b55-b0a4-096ca84748ca",
		Name:         "Sorin's Vengeance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetPlayerOrPlaneswalker(),
		Purpose:      game.Purpose{Targets: game.ForTargets(effects.DamageToTarget(0, 10)), LifeGain: 10, Answers: game.AnswerRemove},
		OnResolve:    effects.ResolveSteps(effects.DamageTheClauseTarget(0, 10), effects.GainLife{Amount: 10}),
	})
}
