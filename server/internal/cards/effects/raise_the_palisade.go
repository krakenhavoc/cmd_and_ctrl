package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raise the Palisade — Sorcery {4}{U}:
//
//	"Choose a creature type. Return all creatures that aren't of the
//	 chosen type to their owners' hands."
//
// Chosen as the spell resolves (#2382); the bounce is one simultaneous
// event (BounceAllMatching). Tokens cease to exist in hand as usual and
// a changeling is of every type, so it stays.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f55a3781-fe33-4301-9bb5-6a54b9c13c4f",
		Name:         "Raise the Palisade",
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			ChooseCreatureTypeThen(ctx.Game, item.Controller, item.SourceCardID, "Raise the Palisade — choose a creature type",
				func(g *game.Game, t string) error {
					if t == "" {
						return nil
					}
					return BounceAllMatching{Match: notOfCreatureType(t)}.Apply(NewContext(g, item))
				})
			return nil
		},
	})
}
