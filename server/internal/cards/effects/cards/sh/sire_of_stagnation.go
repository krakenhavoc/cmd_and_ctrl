package cardssh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sire of Stagnation — Creature — Eldrazi {4}{U}{B}, 5/7:
//
//	"Devoid (This card has no color.)
//	 Whenever a land an opponent controls enters, that player exiles the
//	 top two cards of their library and you draw two cards."
//
// The trigger is Polluted Bonds': it fires for a land an opponent of the
// Sire's controller controls as it enters, and carries that player as its
// parameter. That player exiles the top two cards of their own library
// (face up, a plain exile, CR 701.13), and once the exile has settled
// (CR 608.2c, a replacement or a commander's choice may pause it) the Sire's
// controller draws two cards, even when the library had fewer to exile.
func init() {
	const name = "Sire of Stagnation"
	label := name + " — that player exiles the top two cards of their library and you draw two cards"
	effects.Register(effects.Spec{
		OracleID:        "ed6607f4-20ab-4d3f-a8af-bcc148423336",
		Name:            name,
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{game.KeywordDevoid},
		Triggered: []game.TriggeredAbility{{
			Watches: []game.EventKind{game.EventETB},
			AppliesTo: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				_, ok := effects.B34LandAnOpponentControlsEntered(ev, source, g)
				return ok
			},
			Key: label,
			Build: func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) *game.StackItem {
				land, _ := effects.B34LandAnOpponentControlsEntered(ev, source, g)
				item := game.NewTriggeredItem(source, label)
				item.Params.Player = land.Controller
				return item
			},
			Effect: func(g *game.Game, item *game.StackItem) error {
				ctx := effects.NewContext(g, item)
				victim := item.Params.Player
				if p := g.PlayerByIDForEffect(victim); p == nil || p.Eliminated {
					return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
				}
				return g.ExileTopKeyedForEffect(victim, item.Controller, 2, ctx.ItemTail(sireOfStagnationDraw))
			},
		}},
	})
}

// sireOfStagnationDraw is "and you draw two cards", once the exile has
// landed. An on-disk key.
var sireOfStagnationDraw = effects.CardTail("then/sire-of-stagnation-draw", func(ctx *effects.Context, _ game.TailOutcome) error {
	return effects.DrawCards{Player: ctx.Controller(), N: 2}.Apply(ctx)
})
