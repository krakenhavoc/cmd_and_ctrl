package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Emissary of Soulfire — Creature — Djinn Monk {1}{W}{U}, 1/4:
//
//	"When this creature enters, you get {E}{E}{E} (three energy
//	 counters).
//	 Pay {E}{E}: Put an exalted counter on target creature you control.
//	 Activate only as a sorcery."
//
// #2538 (ADR 0101 amendment 2026-10-08). The exalted counter is a CR
// 122.1b keyword counter, and each one is its own instance of exalted
// (the card's ruling of 2024-06-07), so the engine derives one exalted
// trigger per counter (game/exalted.go). The activation declares no
// Purpose: no field says "put a keyword counter on target creature", so
// the bot prices it as an undeclared activation.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bdacd7a4-3d14-45eb-85a6-33ee711b34af",
		Name:         "Emissary of Soulfire",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Triggered: []game.TriggeredAbility{
			WhenThisEntersYouGetEnergy("Emissary of Soulfire", 3),
		},
		Activated: []ActivatedAbility{{
			Label:        "Pay {E}{E}: Put an exalted counter on target creature you control. Activate only as a sorcery.",
			Cost:         PayEnergy(2),
			SorcerySpeed: true,
			Targets:      TargetCreature("target creature you control", YouControl()),
			Effect:       putACounterOnTheTarget(game.CounterExalted),
		}},
	})
}
