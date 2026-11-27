package cardsex

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Extinction — Sorcery {4}{B} (Tempest, S82 OLD, #5628):
//
//	"Destroy all creatures of the creature type of your choice."
//
// The type is chosen as the spell resolves (#2382), by its controller, the
// way Kindred Dominance chooses. Every creature of that type dies, yours
// included; a changeling is of every type and dies whatever is named.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f3f01fe8-0a2c-4afa-94d2-268d27c0d252",
		Name:         "Extinction",
		Completeness: effects.CompletenessFull,
		// ADR 0126 §6: only creatures of the chosen type die.
		Purpose: game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Extinction — choose a creature type",
				ctx.ItemTail(extinctionTypeChosen))
			return nil
		},
	})
}

// extinctionTypeChosen destroys every creature of the chosen type. An
// on-disk key: never renamed, never reused.
var extinctionTypeChosen = effects.CreatureTypeChosen("then/extinction-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	return effects.DestroyAllMatching{Match: effects.And(effects.Creature(), effects.OfCreatureType(t))}.Apply(ctx)
})
