package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Safe Passage — Instant {2}{W}:
//
//	"Prevent all damage that would be dealt to you and creatures you
//	 control this turn."
//
// ADR 0108 §7, Delivery PR 7 (#1904): the not-one-use shield with no
// source, protecting you and the creatures you control as the damage
// would be dealt (Endure's reading, narrowed to creatures).
//
// No simplifications.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dfa459a1-b065-4488-88d3-4da388261b52",
		Name:         "Safe Passage",
		Completeness: effects.CompletenessFull,
		OnResolve:    effects.SourceShieldSpell(effects.PreventDamageFromSource{Protect: effects.ShieldYouAndYourCreatures}),
	})
}
