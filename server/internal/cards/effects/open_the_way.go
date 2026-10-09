package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Open the Way — Sorcery {X}{G}{G}:
//
//	"X can't be greater than the number of players in the game.
//	 Reveal cards from the top of your library until you reveal X land
//	 cards. Put those land cards onto the battlefield tapped and the
//	 rest on the bottom of your library in a random order."
//
// The first sentence is Spec.XCeiling (#2581): the players are counted
// as X is announced, and only then. Its ruling (2023-05-12): "players
// leaving the game in response to the spell won't affect the number of
// land cards you'll find" — the stack item keeps the announced X.
//
// The reveal is revealUntil with a counting match: it stops on the Xth
// land card, and a library that runs out first reveals itself whole and
// puts every land it found. A token in the library is revealed but is
// not a land card (CR 108.2), so it neither counts nor enters.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e07297c7-83bc-4162-bbb1-362bc737efe0",
		Name:         "Open the Way",
		Completeness: CompletenessFull,
		XMatters:     true,
		XCeiling:     XCeilingPlayersInGame,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return openTheWayReveal(ctx, ctx.X())
		},
	})
}

func openTheWayReveal(ctx *Context, x int) error {
	if x <= 0 {
		return nil
	}
	found := 0
	run, _ := revealUntil(ctx, ctx.Controller(), func(c game.Card) bool {
		if !c.IsLand() {
			return false
		}
		found++
		return found >= x
	}, "Open the Way — revealed until X land cards")
	return PutFromLibraryOntoBattlefield{
		Player: ctx.Controller(),
		Cards:  run,
		Match:  func(_ *game.Game, _ uuid.UUID, c game.Card) bool { return c.IsLand() },
		All:    true,
		Tapped: true,
		Then:   PutRestOnBottomInRandomOrder,
	}.Apply(ctx)
}
