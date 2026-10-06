package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Hindervines — Instant {2}{G}:
//
//	"Prevent all combat damage that would be dealt this turn by
//	 creatures with no +1/+1 counters on them."
//
// #2026's counter test, read as each creature would deal combat damage
// (CR 609.7b; the ruling: "Hindervines checks whether a creature has a
// +1/+1 counter on it at the moment it deals damage").
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "8d13fabc-1e0f-41f2-8873-9f44b68f7e43",
		Name:         "Hindervines",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(combatShieldAgainstCreatures(game.DamageSourceFilter{NoCounters: []string{game.CounterPlusOne}})),
	})
}
