package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harnessed Lightning — Instant {1}{R}:
//
//	"Choose target creature. You get {E}{E}{E} (three energy counters),
//	 then you may pay any amount of {E}. Harnessed Lightning deals that
//	 much damage to that creature."
//
// ADR 0129 §3 (#1995, owner decision 3): the payment is the pay_amount
// prompt, from nothing to every energy the caster has, and the damage is
// the amount paid (CR 118.12). The stepper and the bot start at the
// damage that destroys the creature.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d8dd7e7f-0053-44bb-bac0-1a08843d4c49",
		Name:         "Harnessed Lightning",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Targets:      TargetCreature("target creature"),
		OnResolve:    getEnergyThenPayAnyAmountToDamageTarget("Harnessed Lightning", 3),
	})
}
