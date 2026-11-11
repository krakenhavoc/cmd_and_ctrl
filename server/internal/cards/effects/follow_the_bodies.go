package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Follow the Bodies — Sorcery {2}{U} (Murders at Karlov Manor
// Commander, #5368):
//
//	"Gravestorm (When you cast this spell, copy it for each permanent
//	 put into a graveyard from the battlefield this turn.)
//	 Investigate."
//
// Gravestorm is CR 702.69a, Ominous Harvest's Gravestorm(). The spell
// and each copy investigates once.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2d6a0d89-7ebd-420e-9892-be1506ab3714",
		Name:         "Follow the Bodies",
		Completeness: CompletenessFull,
		Triggered:    []game.TriggeredAbility{Gravestorm()},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return Investigate{N: 1}.Apply(ctx)
		},
	})
}
