package cardsv

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Vine Snare — Instant {2}{G}:
//
//	"Prevent all combat damage that would be dealt this turn by
//	 creatures with power 4 or less."
//
// #2026's power bound, read as each creature would deal combat damage
// (CR 609.7b; the ruling: "It doesn't matter what any creature's power
// is as Vine Snare resolves"). Power counts its counters.
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3303d45-b3ee-4389-8172-d9ffa8bf251b",
		Name:         "Vine Snare",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.SourceShieldSpell(effects.CombatShieldAgainstCreatures(game.DamageSourceFilter{PowerBounded: true, PowerAtMost: 4})),
	})
}
