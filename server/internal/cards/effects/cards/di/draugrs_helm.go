package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Draugr's Helm — Artifact — Equipment {1}{B} (Kaldheim, #5505):
//
//	"When this Equipment enters, you may pay {2}{B}. If you do, create a 2/2
//	 black Zombie Berserker creature token, then attach this Equipment to it.
//	 Equipped creature gets +2/+2 and has menace.
//	 Equip {4}"
//
// Giant's Amulet's shape: the optional payment, then a described token the
// Equipment is attached to (CreateTokenAndAttachSource).
func init() {
	effects.Register(effects.Spec{
		OracleID:     "8762e9ec-e026-45aa-9c7b-d894d17c75a2",
		Name:         "Draugr's Helm",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.PumpAttached(2, 2),
			effects.GrantToAttached("menace"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Draugr's Helm — you may pay {2}{B} to create a 2/2 Zombie Berserker and attach this to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}{B}",
						Question: "Draugr's Helm — pay {2}{B} to create a 2/2 black Zombie Berserker and attach the Helm to it?",
						Then:     draugrsHelmPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{4}")},
	})
}

var draugrsHelmPaid = effects.PayUnlessBranches("pay-unless/draugrs-helm-zombie-and-attach",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateTokenAndAttachSource(ctx.Game, ctx.Item, "2/2 black Zombie Berserker", true)
	}, nil)
