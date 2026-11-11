package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Nymris, Oona's Trickster — Legendary Creature — Faerie Knight
// {3}{U}{B}, 1/6 (Wilds of Eldraine Commander, #5368):
//
//	"Flash
//	 Flying
//	 Whenever you cast your first spell during each opponent's turn, look
//	 at the top two cards of your library. Put one of those cards into
//	 your hand and the other into your graveyard."
//
// The condition is Alela's: a cast of yours during a turn that is not
// yours, and the first one of that turn (the cast tally, which counts
// every spell however it was cast). The look is Faerie Snoop's: a private
// look at the top two, a mandatory pick of one for the hand, the other to
// the graveyard.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9ee50bae-47c4-4ae7-a82b-f2a0b366df35",
		Name:            "Nymris, Oona's Trickster",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flash", "flying"},
		Purpose:         game.Purpose{Draws: 1},
		Triggered: []game.TriggeredAbility{
			On(game.EventCast, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b22FirstSpellOnAnOpponentsTurn(ev, source, g)
			}, "Nymris, Oona's Trickster — look at the top two cards of your library; one into your hand, the other into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
						Max:    1,
						Label:  "Nymris, Oona's Trickster — put one of them into your hand and the other into your graveyard",
						Key:    takeRestGraveyard,
					}.Apply(ctx)
				}),
		},
	})
}
