package cardsgr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Gravkill — Instant {3}{B} (Edge of Eternities, #4461):
//
//	"Exile target creature or Spacecraft."
//
// Angelic Edict's exile of the still-legal target with 70,000
// Light-Years from Home's target clause: "Spacecraft" is the artifact
// subtype, so a Spacecraft below its station threshold (not yet a
// creature) can be exiled too.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "15c6837c-775b-4b4a-a2f7-9fd79215fa05",
		Name:         "Gravkill",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      effects.TargetPermanent("target creature or Spacecraft", effects.Or(effects.Creature(), effects.HasSubtype("Spacecraft"))),
		OnResolve:    effects.ResolveSteps(effects.Instruction(effects.ExileTheFirstTargetIfLegal)),
	})
}
