package cardsst

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Superior Numbers — Sorcery {G}{G} (Mirage, #5633):
//
//	"Superior Numbers deals damage to target creature equal to the number of
//	 creatures you control in excess of the number of creatures target opponent
//	 controls."
//
// Both counts are taken as the spell resolves. It deals nothing if either target
// is illegal by then, or if you control no more creatures than they do.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b98177b5-599f-41b0-aa1e-4769d7005bbd",
		Name:         "Superior Numbers",
		Completeness: effects.CompletenessFull,
		Targets: effects.Clauses(
			effects.TargetCreature("target creature"),
			effects.TargetPlayer("target opponent", effects.Opponent()),
		),
		OnResolve: effects.OldB795SuperiorNumbers,
	})
}
