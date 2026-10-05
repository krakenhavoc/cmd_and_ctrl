package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "ccaa44f2-96be-44e2-884f-c31baa3908d5",
		Name:         "Kindred Dominance",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeThen(ctx.Game, item.Controller, item.SourceCardID, "Kindred Dominance — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					return DestroyAllMatching{Match: notOfCreatureType(t)}.Apply(NewContext(g, item))
				})
			return nil
		},
	})
}
