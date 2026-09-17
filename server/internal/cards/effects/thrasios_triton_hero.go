package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Thrasios, Triton Hero — Legendary Creature — Merfolk Wizard {G}{U},
// 1/3:
//
//	"{4}: Scry 1, then reveal the top card of your library. If it's a
//	 land card, put it onto the battlefield tapped. Otherwise, draw a
//	 card.
//	 Partner"
//
// The mana sink that makes every excess mana a land or a card. The
// ability is a sequence with a real "then": scry 1 is a prompt, and the
// reveal must read the library in the order the player left it, so the
// reveal runs in the scry's continuation (ScryThenForEffect) — the
// Preordain shape. The land half is #745's library-to-battlefield move,
// entering tapped as printed and never counting as a land play
// (CR 305.4); the "otherwise" is an ordinary draw.
//
// A land a replacement keeps off the battlefield is neither put nor
// drawn: "otherwise" is about what the card is.
//
// Partner is a deck-construction rule and needs nothing on the card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3d867016-2601-4a37-a73d-308898d3bd37",
		Name:         "Thrasios, Triton Hero",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label: "{4}: Scry 1, then reveal the top card of your library. If it's a land card, put it onto the battlefield tapped. Otherwise, draw a card.",
			Cost:  ManaCost("{4}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				controller, source := item.Controller, item.SourceCardID
				g.ScryThenForEffect(controller, source, 1, func(g *game.Game) error {
					top, err := revealTopThenPutIfMatch(g, source, controller, game.Card.IsLand, true,
						"Thrasios, Triton Hero — revealed from the top of the library")
					if err != nil {
						return err
					}
					if c, ok := g.LookupCardForEffect(top); ok && c.IsLand() {
						return nil
					}
					return g.DrawNForEffect(controller, 1)
				})
				return nil
			},
		}},
	})
}
