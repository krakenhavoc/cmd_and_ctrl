package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Niblis of the Breath — Creature — Spirit {2}{U} 2/1:
//
//	"Flying
//	 {U}, {T}: You may tap or untap target creature."
func init() {
	effects.Register(effects.Spec{
		OracleID:        "10391bcb-0932-49a0-8d47-e06e3c10dd68",
		Name:            "Niblis of the Breath",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{U}, {T}: You may tap or untap target creature.",
			Cost:    game.AbilityCost{Tap: true, Mana: "{U}"},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.NemMayTapOrUntapTheTarget("Niblis of the Breath"),
		}},
	})
}
