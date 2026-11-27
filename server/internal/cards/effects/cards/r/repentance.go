package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Repentance — Sorcery {2}{W} (Tempest, S82 OLD batch 758, #5628):
//
//	"Target creature deals damage to itself equal to its power."
//
// Justice Strike's body: the creature is its own damage source and its power
// is read as the spell resolves.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "9a345a67-0c6c-4f6f-b3c9-d525199cdbfa",
		Name:         "Repentance",
		Completeness: effects.CompletenessFull,
		Targets:      effects.TargetCreature("target creature"),
		OnResolve:    effects.B583TargetCreatureDamagesItself,
	})
}
