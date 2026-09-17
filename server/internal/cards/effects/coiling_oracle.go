package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Coiling Oracle — Creature — Snake Elf Druid {G}{U}, 1/1:
//
//	"When this creature enters, reveal the top card of your library.
//	 If it's a land card, put it onto the battlefield. Otherwise, put
//	 that card into your hand."
//
// A two-mana creature that is always a card and sometimes a land. The
// land half is #745's library-to-battlefield move: the reveal is not a
// search (no search payoff sees it, nothing shuffles), and the land is
// PUT, not played, so the turn's land drop survives it (CR 305.4) and
// its own enters-tapped clause still applies through the CR 614
// pipeline. The "otherwise" half is an ordinary move to hand, which is
// not a draw.
//
// A land that a replacement keeps off the battlefield stays on top of
// the library: the card put nothing onto the battlefield, and
// "otherwise" is about what the revealed card IS, not about whether
// the put worked.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "69fd4ddf-9ed8-4c56-bef3-9944daf05e4f",
		Name:         "Coiling Oracle",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			WhenThisEnters("Coiling Oracle — reveal the top card; a land goes onto the battlefield, anything else into your hand",
				func(g *game.Game, item *game.StackItem) error {
					top, err := revealTopThenPutIfMatch(g, item.SourceCardID, item.Controller, game.Card.IsLand, false,
						"Coiling Oracle — revealed from the top of the library")
					if err != nil || top == uuid.Nil {
						return err
					}
					if c, ok := g.LookupCardForEffect(top); ok && (c.IsLand() || IsToken(c)) {
						// A token tucked into the library earlier is
						// neither put nor moved to hand: it can't
						// change zones again (CR 111.8).
						return nil
					}
					return g.BounceToHandForEffect(top)
				}),
		},
	})
}
