package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Galvanic Discharge — Instant {R}:
//
//	"Choose target creature or planeswalker. You get {E}{E}{E} (three
//	 energy counters), then you may pay any amount of {E}. Galvanic
//	 Discharge deals that much damage to that permanent."
//
// ADR 0129 §3 (#1995, owner decision 3): Harnessed Lightning's sentence
// with a planeswalker allowed. The stepper and the bot start at the
// damage that destroys the permanent: a creature's toughness less its
// marked damage, a planeswalker's loyalty (CR 120.3c).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "ef0f06f1-3991-4897-bf62-d4cab7ec79a0",
		Name:         "Galvanic Discharge",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Energy: 3},
		Targets:      effects.TargetPermanent("target creature or planeswalker", effects.Or(effects.Creature(), effects.Planeswalker())),
		OnResolve:    effects.GetEnergyThenPayAnyAmountToDamageTarget("Galvanic Discharge", 3),
	})
}
