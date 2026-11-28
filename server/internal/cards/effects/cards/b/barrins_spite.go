package cardsb

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Barrin's Spite — Sorcery {2}{U}{B} (Invasion, S82 older sets batch 662, #5616):
//
//	"Choose two target creatures controlled by the same player. Their
//	 controller chooses and sacrifices one of them. Return the other to its
//	 owner's hand."
//
// Bioshift's pair of related targets (the second controlled by the first's
// controller, ControlledByTargetsController, and distinct from it). Their
// controller as the spell resolves picks which one to sacrifice; the other is
// bounced. If one target is illegal, the other is the one sacrificed.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "68b038ef-71c4-4ca5-9821-57bf85831bb7",
		Name:         "Barrin's Spite",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.Distinct(effects.ControlledByTargetsController(effects.TargetCreature("another target creature controlled by the same player"), 0)),
		),
		OnResolve: effects.OldB662BarrinsSpite,
	})
}
