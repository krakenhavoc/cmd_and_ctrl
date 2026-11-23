package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Wanderer's Strike — Sorcery {4}{W} (Commander Masters, #5948):
//
//	"Exile target creature, then proliferate. (Choose any number of
//	 permanents and/or players, then give each another counter of each kind
//	 already there.)"
//
// Blade Banish's exile of the target (only while it is still legal, CR
// 608.2b), then the player's proliferate choice (Atomize's), which still
// happens when the exile did nothing to an indestructible-proof target.
// With its only target gone the spell does not resolve at all, so there is
// no proliferate either.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d6239ada-c72a-49b2-882c-6dae1dd366e3",
		Name:         "Wanderer's Strike",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.ResolveSteps(effects.ExileChosenTarget{}, effects.Proliferate{}),
	})
}
