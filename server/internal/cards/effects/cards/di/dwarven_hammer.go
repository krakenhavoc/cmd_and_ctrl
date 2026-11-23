package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Dwarven Hammer — Artifact — Equipment {2}{R} (Commander Masters):
//
//	"When this Equipment enters, you may pay {2}. If you do, create a 2/1
//	 red Dwarf Berserker creature token, then attach this Equipment to it.
//	 Equipped creature gets +3/+0 and has trample.
//	 Equip {3}"
//
// Giant's Amulet's enter trigger: the optional {2} is MayPay on resolution,
// and the attach rides the token's creation (createTokenAndAttachSource),
// so if the Hammer has left in response the Dwarf is made and nothing is
// attached.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a6365f98-f652-492b-81de-5f8cd6fe0a1c",
		Name:         "Dwarven Hammer",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(3, 0), effects.GrantToAttached("trample")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Dwarven Hammer — you may pay {2} to create a 2/1 red Dwarf Berserker and attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Dwarven Hammer — pay {2} to create a 2/1 red Dwarf Berserker and attach the Hammer to it?",
						Then:     dwarvenHammerPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

// dwarvenHammerPaid is the paid half, keyed: the Dwarf, and the attach to the
// Hammer the asking item names.
var dwarvenHammerPaid = effects.PayUnlessBranches("pay-unless/dwarven-hammer-dwarf-and-attach",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateTokenAndAttachSource(ctx.Game, ctx.Item, "2/1 red Dwarf Berserker", true)
	}, nil)
