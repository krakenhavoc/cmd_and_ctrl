package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dismissive Denial — Instant {2}{U}{U} (Marvel Super Heroes Commander,
// #4455): "Counter target spell. Basic landcycling {2}". Cancel's
// resolution (counterTheTargetSpell) plus the cycling family's
// BasicLandcycling.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "99eff6e9-5911-4f89-bad1-0888067a7783",
		Name:         "Dismissive Denial",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetSpell("target spell"),
		OnResolve:    effects.CounterTheTargetSpell,
		Activated:    []effects.ActivatedAbility{effects.BasicLandcycling("{2}")},
	})
}
