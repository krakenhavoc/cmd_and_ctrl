package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Kindred Judgment — Sorcery {5}{W}{W}:
//
//	"Choose a creature type. Destroy all creatures that aren't of the
//	 chosen type."
//
// The type is chosen by the caster as the spell resolves (#2382). A
// changeling is of every type and survives; a creature that is not of
// the type dies, yours included, and regeneration and indestructible
// apply as for any destroy effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b5dce42a-a769-4d7d-b29e-8722f79d4092",
		Name:         "Kindred Judgment",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *effects.Context) error {
			effects.ChooseCreatureTypeTail(ctx.Game, item.Controller, item.SourceCardID, "Kindred Judgment — choose a creature type",
				ctx.ItemTail(kindredJudgmentTypeChosen))
			return nil
		},
	})
}

// kindredJudgmentTypeChosen destroys every creature not of the chosen type.
// An on-disk key.
var kindredJudgmentTypeChosen = effects.CreatureTypeChosen("then/kindred-judgment-type-chosen", func(ctx *effects.Context, t string, _ game.TailOutcome) error {
	return effects.DestroyAllMatching{Match: effects.Except(effects.Creature(), effects.OfCreatureType(t))}.Apply(ctx)
})
