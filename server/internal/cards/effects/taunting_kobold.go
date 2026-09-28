package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Taunting Kobold — Creature — Kobold, {R}, 0/1:
//
//	"Haste
//	 Whenever this creature attacks, goad target creature an
//	 opponent controls. (Until your next turn, that creature attacks
//	 each combat if able and attacks a player other than you if
//	 able.)"
//
// #1599: haste rides PrintedKeywords; the attack trigger is GoadTarget
// (goad.go) over OpponentControls() — "an opponent controls" reads
// against the trigger's own controller (the attacking Kobold's), the
// same caster argument every other Targets clause on this ability
// gets.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "604c581e-c1ba-417c-b9de-182192cc85cb",
		Name:            "Taunting Kobold",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Triggered: []game.TriggeredAbility{{
			Watches:   []game.EventKind{game.EventAttack},
			AppliesTo: ThisAttacked,
			Targets:   TargetCreature("target creature an opponent controls", OpponentControls()),
			Key:       "Taunting Kobold — goad target creature an opponent controls",
			Effect:    GoadTarget,
		}},
	})
}
