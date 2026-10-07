package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Longtusk Cub — Creature — Cat {1}{G}, 2/2:
//
//	"Whenever this creature deals combat damage to a player, you get
//	 {E}{E} (two energy counters).
//	 Pay {E}{E}: Put a +1/+1 counter on this creature."
//
// ADR 0129 PR 1 (#1995).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d2e78392-096a-4bd0-be19-9c54dfb8451f",
		Name:         "Longtusk Cub",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WheneverThisDealsCombatDamageToAPlayer("Longtusk Cub — you get {E}{E}", ebYouGetEnergy(2)),
		},
		Activated: []ActivatedAbility{{
			Label:  "Pay {E}{E}: Put a +1/+1 counter on this creature.",
			Cost:   PayEnergy(2),
			Effect: plusOneCountersOnThis(1),
		}},
	})
}
