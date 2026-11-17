package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Pulverize — Sorcery {4}{R}{R} (Mercadian Masques, #5619):
//
//	"You may sacrifice two Mountains rather than pay this spell's mana
//	 cost.
//	 Destroy all artifacts."
//
// Fireblast's two-Mountain alternative cost (SacrificeInstead, #1727) on
// Shatterstorm's sweep, without Shatterstorm's "can't be regenerated":
// a regeneration shield on an artifact still saves it here (CR 701.19a).
// Every artifact on the battlefield as it resolves, the caster's
// included.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d403d1f3-5a92-4f4c-aff2-cff6a0b3b1ec",
		Name:         "Pulverize",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepArtifacts, How: game.SweepDestroy}},
		AlternativeCosts: []game.AlternativeCost{
			SacrificeInstead(2, "two Mountains", 0, HasSubtype("Mountain")),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return DestroyAllMatching{Match: Artifact()}.Apply(ctx)
		},
	})
}
