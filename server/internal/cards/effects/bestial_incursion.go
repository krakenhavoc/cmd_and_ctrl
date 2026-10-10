package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bestial Incursion — Sorcery {3}{G}:
//
//	"Create a 4/4 green Beast creature token with trample.
//	 Flashback {5}{G} (You may cast this card from your graveyard for
//	 its flashback cost. Then exile it.)"
//
// Flashback is the shared constructor, which bundles the graveyard
// cast zone, the price and CR 702.34a's exile-on-leaving-the-stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:         "bc00f29a-e9d2-4b86-bb22-33b686ac4365",
		Name:             "Bestial Incursion",
		Completeness:     CompletenessFull,
		CastableZones:    []game.ZoneKind{game.ZoneGraveyard},
		AlternativeCosts: []game.AlternativeCost{Flashback("{5}{G}")},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("4/4 green Beast with trample"),
				N:          1,
			}.Apply(ctx)
		},
	})
}
