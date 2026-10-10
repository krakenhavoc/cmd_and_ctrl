package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "b5dce42a-a769-4d7d-b29e-8722f79d4092",
		Name:         "Kindred Judgment",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Sweep: game.Sweep{Matches: game.SweepCreatures, How: game.SweepDestroy, Partial: true}},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeThen(ctx.Game, item.Controller, item.SourceCardID, "Kindred Judgment — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					return DestroyAllMatching{Match: Except(Creature(), OfCreatureType(t))}.Apply(NewContext(g, item))
				})
			return nil
		},
	})
}
