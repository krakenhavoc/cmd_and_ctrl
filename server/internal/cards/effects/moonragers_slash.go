package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Moonrager's Slash — {2}{R} Instant (#2561, ADR 0132):
//
//	"This spell costs {2} less to cast if it's night.
//	 Moonrager's Slash deals 3 damage to any target."
//
// The discount reads the designation as the spell is priced, so it is
// there on the turn night begins and gone the moment it is day again.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "40a41774-fe50-438d-85e4-7f2beb29fb82",
		Name:         "Moonrager's Slash",
		Completeness: CompletenessFull,
		Targets:      TargetAny(),
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "This spell costs {2} less to cast if it's night.", ItsNightCost()),
		},
		OnResolve: damageToFirstTarget(3),
	})
}
