package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Braingeyser — Sorcery {X}{U}{U} (Limited Edition Alpha):
//
//	"Target player draws X cards."
//
// The shared X draw (TargetPlayerDrawsX), as Dirgur Focusmage's prepare
// spell of the same name prints it. Built with Garth One-Eye (#6155),
// whose copy of "the card with the chosen name" resolves through this
// entry.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9908e597-9470-4c13-8387-39431b380138",
		Name:         "Braingeyser",
		Completeness: effects.CompletenessFull,
		XMatters:     true,
		Targets:      effects.TargetPlayer("target player"),
		OnResolve:    effects.TargetPlayerDrawsX(),
	})
}
