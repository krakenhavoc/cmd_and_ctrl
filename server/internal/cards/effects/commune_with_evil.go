package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Commune with Evil — Sorcery {2}{B} (Duskmourn: House of Horror, #4466):
//
//	"Look at the top four cards of your library. Put one of them into
//	 your hand and the rest into your graveyard. You gain 3 life."
//
// Pattern Buffer Preservation's look-and-take over four cards: a private
// look, a mandatory pick of one for the hand, and the others put into the
// graveyard (not milled or discarded). The life comes after the cards
// have moved, in the take's registered continuation
// (communeWithEvilTake), so it waits for the pick.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0248bf01-14e0-4a39-ba3a-5177fe272fe9",
		Name:         "Commune with Evil",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Draws: 1},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			player := ctx.Controller()
			return TakeFromLibraryToHand{
				Player: player,
				Cards:  ctx.Game.LookAtTopOfLibraryForEffect(player, 4),
				Max:    1,
				Label:  "Commune with Evil — put one of them into your hand and the rest into your graveyard",
				Key:    communeWithEvilTake,
			}.Apply(ctx)
		},
	})
}

// communeWithEvilTake is the look's continuation: the rest into the
// graveyard, then "you gain 3 life".
var communeWithEvilTake = RegisterTake("choose-cards/commune-with-evil-take",
	TakeDef{Then: func(g *game.Game, res TakeFromLibraryResult) error {
		if err := TakeRestIntoGraveyard(g, res); err != nil {
			return err
		}
		return g.ChangePlayerLifeForEffect(res.Source, res.Player, 3)
	}})
