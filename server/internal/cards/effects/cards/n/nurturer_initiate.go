package cardsn

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Nurturer Initiate — Creature — Elf Shaman {G} 1/1:
//
//	"Whenever a player casts a green spell, you may pay {1}. If you do, target
//	 creature gets +1/+1 until end of turn."
//
// Any player's green spell triggers it; the target is chosen as the trigger
// goes on the stack (CR 603.3d) and the {1} is offered as it resolves
// (Battlemage's Bracers' offer). The pump is applied only if the payment is
// made and the target is still legal (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "78064301-f81e-4737-9cfe-b1e8757d0dd6",
		Name:         "Nurturer Initiate",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Targeting(effects.On(game.EventCast, func(ev game.Event, _ *game.Card, _ game.Characteristic, g *game.Game) bool {
				spell, ok := g.LookupCardForEffect(ev.CardID)
				return ev.Kind == game.EventCast && ok && spell.HasColorWord("G")
			}, "Nurturer Initiate — you may pay {1}; if you do, target creature gets +1/+1 until end of turn",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Nurturer Initiate — pay {1} to give target creature +1/+1 until end of turn?",
						Then:     nurturerInitiatePaid,
					}.Apply(effects.NewContext(g, item))
				}),
				effects.TargetCreature("target creature")),
		},
	})
}

var nurturerInitiatePaid = effects.PayUnlessBranches("pay-unless/nurturer-initiate-pump",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		id, ok := effects.B16FirstLegalTargetCard(ctx)
		if !ok {
			return nil
		}
		return effects.BoostUntilEOT{Target: id, Power: 1, Toughness: 1, Label: "Nurturer Initiate — +1/+1 until end of turn"}.Apply(ctx)
	}, nil)
