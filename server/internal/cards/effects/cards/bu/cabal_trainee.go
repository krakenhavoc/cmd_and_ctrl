package cardsbu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Cabal Trainee — Creature — Human Minion {B}, 1/1 (Judgment):
//
//	"Sacrifice this creature: Target creature gets -2/-0 until end of turn."
//
// A sacrifice-cost shrink: the target gets -2/-0 until end of turn if it is
// still legal as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "fce68545-c31d-4dc2-9f1b-aadf924a07d1",
		Name:         "Cabal Trainee",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice this creature: Target creature gets -2/-0 until end of turn.",
			Purpose: effects.WithAnswers(effects.ForTargets(effects.HarmUntilEOT(0, -2, 0)), game.AnswerPump),
			Cost:    effects.SacrificeThis(),
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.BoostTargetUntilEOT("Cabal Trainee — -2/-0", -2, 0),
		}},
	})
}
