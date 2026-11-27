package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Agonizing Demise (Invasion, S82 OLD batch 648, #5616):
//
//	Instant {3}{B}:
//	"Kicker {1}{R}
//	Destroy target nonblack creature. It can't be regenerated. If this spell was kicked,
//	Agonizing Demise deals damage equal to that creature's power to the creature's controller."
func init() {
	effects.Register(effects.Spec{
		OracleID:      "41681a1f-0777-4b96-8c7e-d6a5d1b1c8fc",
		Name:          "Agonizing Demise",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Answers: game.AnswerRemove},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{1}{R}")},
		Targets:       effects.TargetCreature("target nonblack creature", effects.NonBlack()),
		OnResolve:     effects.S82OldB648AgonizingDemiseResolve,
	})
}
