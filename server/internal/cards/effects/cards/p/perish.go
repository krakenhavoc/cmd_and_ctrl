package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Perish — Sorcery {2}{B} (Tempest, S82 OLD, #5628):
//
//	"Destroy all green creatures. They can't be regenerated."
//
// Virtue's Ruin's sweep turned to green, with Wrath of God's rider: a
// creature with a regeneration shield is destroyed anyway (CR 701.19c).
// Green is read through the layered colour, so a creature that has become
// green dies and one that has lost the colour stays.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "dd84d291-cb7a-4f44-81cd-9f1181bb5ae5",
		Name:         "Perish",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: a creature that isn't green stays.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.OfColor("G")), CantBeRegenerated: true}.Apply(ctx)
		},
	})
}
