package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Rise of the Dark Realms — Sorcery {7}{B}{B} (EDHREC rank 501):
//
//	"Put all creature cards from all graveyards onto the battlefield
//	 under your control."
//
// Nine mana to take every dead creature at the table. Every
// graveyard is walked, the caster's included, and each creature
// card returns UNDER THE CASTER'S CONTROL — the Controller field on
// ReturnFromGraveyardTogether, which is what makes an opponent's dead
// bomb yours rather than theirs (the Reanimate lesson). The cards are
// read before any of them moves, and they enter together, as one
// event (#1867): every returned creature fires its own ETB, and each
// sees the others enter (CR 603.6a).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e5223a09-f732-4747-8914-e6546ab0ef4c",
		Name:         "Rise of the Dark Realms",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{
				Targets:    allGraveyardsCardIDs(ctx, game.Card.IsCreature),
				Controller: ctx.Controller(),
			}.Apply(ctx)
		},
	})
}
