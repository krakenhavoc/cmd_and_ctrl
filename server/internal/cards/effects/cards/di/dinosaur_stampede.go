package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dinosaur Stampede — Instant {2}{R} (Ixalan):
//
//	"Attacking creatures get +2/+0 until end of turn. Dinosaurs you control
//	 gain trample until end of turn."
//
// Morale's sweep for the pump, then a keyword sweep over the Dinosaurs its
// controller controls. Both sets are fixed as the spell resolves (CR 611.2c),
// so a creature that arrives or starts attacking later gets nothing.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "3d6af1ee-16d2-4b82-8d9f-6d956603a746",
		Name:         "Dinosaur Stampede",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			if err := (effects.BoostUntilEOT{Match: effects.AttackingCreature(), Power: 2, Label: "Dinosaur Stampede — attacking creatures get +2/+0"}).Apply(ctx); err != nil {
				return err
			}
			return effects.GrantKeywordUntilEOT{
				Match:    effects.And(effects.Creature(), effects.Subtype("Dinosaur"), effects.YouControl()),
				Keywords: []string{"trample"},
				Label:    "Dinosaur Stampede — Dinosaurs you control gain trample",
			}.Apply(ctx)
		},
	})
}
