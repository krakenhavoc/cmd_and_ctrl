package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Armory — Enchantment {2}{R}{R} (Ice Age, S82 OLD batch 845, #5636):
//
//	"{2}, Sacrifice a land: Put a +2/+2 counter on target creature. Activate
//	 only during any upkeep step."
//
// The sacrifice is a cost paid at announce (CR 118.3); "any upkeep step" is
// DuringStep(StepUpkeep), whoever's upkeep it is (Armageddon Clock's
// instruction about activating, CR 602.1b). The +2/+2 counter is a counter
// kind the engine already reads as +2/+2 (Chimera Sacrifice, pt_counters.go).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "aba53a54-8666-430d-899e-45ad9f783a4a",
		Name:         "Dwarven Armory",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:     "{2}, Sacrifice a land: Put a +2/+2 counter on target creature. Activate only during any upkeep step.",
			Purpose:   game.Purpose{Answers: game.AnswerPump},
			Cost:      effects.Plus(effects.ManaCost("{2}"), effects.B08SacrificeALand()),
			Targets:   effects.TargetCreature("target creature"),
			Condition: effects.DuringStep(game.StepUpkeep),
			Effect:    effects.PutACounterOnTheTarget("+2/+2"),
		}},
	})
}
