package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Wear Away — Instant — Arcane {G}{G}:
//
//	"Destroy target artifact or enchantment.
//	Splice onto Arcane {3}{G}"
func init() {
	Register(Spec{
		OracleID:     "be7cd599-12b6-43b6-a22a-8a8f7dd1ecd2",
		Name:         "Wear Away",
		Completeness: CompletenessFull,
		Splice:       SpliceOntoArcane("{3}{G}"),
		Targets:      TargetPermanent("target artifact or enchantment", Or(Artifact(), Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			t, ok := firstLegalTarget(ctx)
			if !ok || t.Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: t.ID}.Apply(ctx)
		},
	})
}
