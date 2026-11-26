package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Touch of Moonglove — Instant {B} (Magic Origins, S82 older sets):
//
//	"Target creature you control gets +1/+0 and gains deathtouch until end of
//	 turn. Whenever a creature dealt damage by that creature dies this turn,
//	 its controller loses 2 life. (Any amount of damage a creature with
//	 deathtouch deals to a creature is enough to destroy it.)"
//
// Rushed Rebirth's shape: the pump and deathtouch are fixed as the spell
// resolves, and a "whenever … this turn" delayed trigger (CR 603.7b) about the
// target as the object it is now (CR 603.7c) watches every creature that dies.
// A creature counts only if the target dealt damage to it this turn (combat or
// not, prevented damage never counts); its controller, read as it last
// existed, loses 2 life for each. A target gone by resolution makes the spell
// do nothing (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d3a6bff6-ea85-4e8a-b256-7427d15326b6",
		Name:         "Touch of Moonglove",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature you control", effects.YouControl()),
		Purpose:      game.Purpose{Answers: game.AnswerPump, Targets: game.ForTargets(game.TargetPurpose{Slot: 0, UntilEndOfTurn: game.TargetBenefit})},
		OnResolve:    effects.OldB321TouchOfMoonglove("Touch of Moonglove"),
	})
}
