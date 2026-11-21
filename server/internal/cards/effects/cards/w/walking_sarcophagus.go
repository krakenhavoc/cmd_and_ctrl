package cardsw

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Walking Sarcophagus — Artifact Creature — Zombie Cat {2}, 2/1:
//
//	"Start your engines!
//	 Max speed — This creature gets +1/+2."
//
// ADR 0138 (#2711).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "791b8ac8-8db8-44b4-8c85-fcc9017c17ef",
		Name:            "Walking Sarcophagus",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{effects.StartYourEngines},
		Static:          []game.StaticAbility{effects.MaxSpeedSelfPump(1, 2)},
	})
}
