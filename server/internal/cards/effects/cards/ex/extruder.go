package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Extruder — Artifact Creature — Juggernaut, {4}, 4/3:
//
//	"Echo {4} (At the beginning of your upkeep, if this came under your control since the beginning of your last upkeep, sacrifice it unless you pay its echo cost.)
//	 Sacrifice an artifact: Put a +1/+1 counter on target creature."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "93cc6425-a13b-4b32-bb8f-fdc9bfa0ef9b",
		Name:         "Extruder",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "Sacrifice an artifact: Put a +1/+1 counter on target creature.",
			Purpose: game.Purpose{Answers: game.AnswerPump | game.AnswerSacOutlet},
			Cost:    game.AbilityCost{SacrificeOther: effects.SacrificeSpec("an artifact", effects.Artifact())},
			Targets: effects.TargetCreature("target creature"),
			Effect:  effects.PutPlusOneCounterOnEachLegalTarget,
		}},
		Triggered: []game.TriggeredAbility{effects.Echo("Extruder", "{4}")},
	})
}
