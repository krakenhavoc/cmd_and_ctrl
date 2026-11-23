package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hammer Dropper — Creature — Giant Soldier {2}{R}{W}, 5/2:
//
//	"Mentor (Whenever this creature attacks, put a +1/+1 counter on
//	 target attacking creature with lesser power.)"
//
// Mentor (CR 702.134) is the shared Mentor() trigger: its clause is
// relative to the mentor (#2146), so the target's power is compared
// with the mentor's as the trigger goes on the stack and again as it
// resolves, and the trigger is dropped when no attacker has lesser
// power (CR 603.3d).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "088c9fa7-65ec-48c6-90cc-9087bc9df43e",
		Name:         "Hammer Dropper",
		Completeness: effects.CompletenessFull,
		Triggered:    []game.TriggeredAbility{effects.Mentor("Hammer Dropper — mentor")},
	})
}
