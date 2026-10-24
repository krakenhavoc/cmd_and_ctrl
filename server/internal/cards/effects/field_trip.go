package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Field Trip — Sorcery {2}{G} (EDHREC rank 18535):
//
//	"Search your library for a basic Forest card, put that card onto
//	 the battlefield tapped, then shuffle.
//	 Learn. (You may reveal a Lesson card you own from outside the game
//	 and put it into your hand, or discard a card to draw a card.)"
//
// The search asks for a pick, so learn waits for it as a later step
// (CR 608.2c); a player is not asked what to discard before the Forest
// is chosen and the library shuffled.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "29423b70-7cb7-4617-898c-78a38eb233ad",
		Name:         "Field Trip",
		Purpose:      game.Purpose{Lands: 1},
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(
			Instruction(func(ctx *Context) error {
				return SearchLibrary{
					Player:        ctx.Controller(),
					Predicate:     IsBasicLandWithSubtype("forest"),
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Field Trip — a basic Forest card",
				}.Apply(ctx)
			}),
			Learn{},
		),
	})
}
