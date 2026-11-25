package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Miara, Thorn of the Glade — Legendary Creature — Elf Scout {1}{B}, 1/2
// (Kaldheim Commander, #5506):
//
//	"Whenever Miara or another Elf you control dies, you may pay {1} and 1
//	 life. If you do, draw a card.
//	 Partner"
//
// "Elf you control" is read as the creature last existed (CR 603.10a),
// Miara herself included. The optional cost is Ripples of Undeath's split:
// the {1} rides the may-pay prompt and the life is paid through the cost
// path inside the paid branch, with the offer withheld below 1 life (CR
// 119.4). Partner is a deck-construction rule and needs nothing on the card.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "6895d7a2-264b-4731-9294-c02e1e220384",
		Name:         "Miara, Thorn of the Glade",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.S82OldB074ElfYouControlDied(ev, source, g)
			}, "Miara, Thorn of the Glade — you may pay {1} and 1 life; if you do, draw a card",
				func(g *game.Game, item *game.StackItem) error {
					p := g.PlayerByIDForEffect(item.Controller)
					if p == nil || p.Life < 1 {
						return nil
					}
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Miara — pay {1} and 1 life to draw a card?",
						Then:     miaraPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

// miaraPaid is the paid branch: the life, then the card. An on-disk key:
// never renamed, never reused.
var miaraPaid = effects.PayUnlessBranches("pay-unless/miara-draw",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		if err := ctx.Game.PayLifeForEffect(ctx.Source(), ctx.Controller(), 1); err != nil {
			return err
		}
		return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
	}, nil)
