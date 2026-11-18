package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Warm Welcome — Instant {2}{G} (Streets of New Capenna):
//
//	"Look at the top five cards of your library. You may reveal a creature
//	 card from among them and put it into your hand. Put the rest on the
//	 bottom of your library in a random order. Create a 1/1 green and white
//	 Citizen creature token."
//
// Commune with Nature's dig over creature cards, the rest going to the
// bottom in a random order, then the Citizen; the token is made whether or
// not a card was taken.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53807ded-5bbb-44a1-9c22-ffdd4b8523e0",
		Name:         "Warm Welcome",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			if err := (TakeFromLibraryToHand{
				Player:   player,
				Cards:    ctx.Game.LookAtTopOfLibraryForEffect(player, 5),
				Match:    Creature(),
				Max:      1,
				Optional: true,
				Label:    "Warm Welcome — reveal a creature card and put it into your hand",
				Key:      takeRevealRestBottomRandom,
			}).Apply(ctx); err != nil {
				return err
			}
			return CreateToken{Template: TokenCard("1/1 green and white Citizen"), N: 1}.Apply(ctx)
		},
	})
}
