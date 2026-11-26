package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// End Hostilities — Sorcery {3}{W}{W} (Khans of Tarkir):
//
//	"Destroy all creatures and all permanents attached to creatures."
//
// One simultaneous destruction (CR 701.8, Wrath of God's sweep) of every
// creature and every permanent attached to a creature, judged as the spell
// resolves: Equipment and Auras on the creatures go with them, a Fortification
// or an Equipment attached to something that is not a creature stays, and
// regeneration may replace any of it.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "7b2e2c4b-b8fd-440c-8a79-de6532022745",
		Name:         "End Hostilities",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: every creature, and what is attached to one.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.OldB352CreatureOrAttachedToOne}.Apply(ctx)
		},
	})
}
