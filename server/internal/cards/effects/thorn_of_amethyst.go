package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thorn of Amethyst — Artifact {2}:
//
//	"Noncreature spells cost {1} more to cast."
//
// Sphere of Resistance with a type predicate. Still symmetrical — it
// taxes its controller's removal spell exactly as it taxes everyone
// else's.
func init() {
	Register(Spec{
		OracleID:     "0c6c5336-9233-4ab1-9d55-79f20be7ea57",
		Name:         "Thorn of Amethyst",
		Completeness: CompletenessFull,
		CostModifiers: []game.CostModifier{
			CostsMore(1, "Noncreature spells cost {1} more to cast.", NoncreatureSpell()),
		},
	})
}
