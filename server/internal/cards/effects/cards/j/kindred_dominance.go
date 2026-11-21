package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kindred Dominance — Sorcery {5}{B}{B}:
//
//	"Choose a creature type. Destroy all creatures that aren't of the
//	 chosen type."
//
// Chosen as the spell resolves (#2382). Every creature on the table is
// swept, yours included; a changeling is of every type and survives
// whatever is named. Indestructible creatures survive the destroy as
// usual.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID: "ccaa44f2-96be-44e2-884f-c31baa3908d5",
		Name:     "Kindred Dominance",
		// ADR 0126 §6: creatures of the chosen type stay.
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		Completeness: effects.CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Kindred Dominance — choose a creature type",
				ctx.ItemTail(kindredDominanceTypeChosen))
			return nil
		},
	})
}

// kindredDominanceTypeChosen destroys every creature not of the chosen
// type. An on-disk key.
var kindredDominanceTypeChosen = effects.CreatureTypeChosen("then/kindred-dominance-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	return effects.DestroyAllMatching{Match: effects.NotOfCreatureType(t)}.Apply(ctx)
})
