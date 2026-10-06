package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Harmless Assault — Instant {2}{W}{W}:
//
//	"Prevent all combat damage that would be dealt this turn by
//	 attacking creatures."
//
// #2026's combat status, read as each creature would deal combat
// damage (CR 609.7b): an attacking creature's combat damage is
// prevented, and a blocker's is not (the ruling).
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "bfcc3a16-1ca8-4112-a10e-d4e52d3daa8d",
		Name:         "Harmless Assault",
		Completeness: CompletenessFull,
		OnResolve:    sourceShieldSpell(combatShieldAgainstCreatures(game.DamageSourceFilter{Combat: game.SourceCombatAttacking})),
	})
}
