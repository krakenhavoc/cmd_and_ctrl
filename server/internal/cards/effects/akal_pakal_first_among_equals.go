package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Akal Pakal, First Among Equals — Legendary Creature — Human Advisor
// {2}{U}, 1/5 (The Lost Caverns of Ixalan, #4470):
//
//	"At the beginning of each player's end step, if an artifact entered
//	 the battlefield under your control this turn, look at the top two
//	 cards of your library. Put one of them into your hand and the other
//	 into your graveyard."
//
// It triggers in every player's end step, not only yours. The condition
// is an intervening if (CR 603.4) on Master's Manufactory's turn tally
// (game.EnteredWithCardTypeThisTurn), recorded as each permanent entered,
// so an artifact that has since left or stopped being an artifact still
// counts. A token artifact counts too. The look is Orazca Puzzle-Door's:
// a private look at the top two, a mandatory pick of one for the hand,
// and the other put into the graveyard (takeRestGraveyard).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "53bfe54e-9b97-4113-97d9-3c5dbc85e7cf",
		Name:         "Akal Pakal, First Among Equals",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			InterveningIf(CondBoard{
				Name: "an artifact entered the battlefield under your control this turn",
				Test: func(ctx *Context) bool { return ctx.Game.EnteredWithCardTypeThisTurn(ctx.Controller(), "artifact") > 0 },
			}, On(game.EventBeginEndStep, func(_ game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return g.EnteredWithCardTypeThisTurn(source.Controller, "artifact") > 0
			}, "Akal Pakal — look at the top two cards; one into your hand, the other into your graveyard",
				func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					player := ctx.Controller()
					return TakeFromLibraryToHand{
						Player: player,
						Cards:  g.LookAtTopOfLibraryForEffect(player, 2),
						Max:    1,
						Label:  "Akal Pakal — put one of them into your hand and the other into your graveyard",
						Key:    takeRestGraveyard,
					}.Apply(ctx)
				})),
		},
	})
}
