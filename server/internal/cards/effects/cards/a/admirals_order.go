package cardsa

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Admiral's Order — Instant {1}{U}{U} (Rivals of Ixalan, S82 OLD batch 218,
// #5531):
//
//	"Raid — If you attacked this turn, you may pay {U} rather than pay this
//	 spell's mana cost.
//	 Counter target spell."
//
// Cancel with an alternative cost (CR 118.9) offered only to a player who has
// declared an attack this turn, checked as the spell is announced and in the
// view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "ae4fe5e0-408d-4398-84d7-a901b836df2a",
		Name:             "Admiral's Order",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.RaidPay("{U}")},
		Targets:          effects.TargetSpell("target spell"),
		OnResolve:        effects.CounterTheTargetSpell,
	})
}
