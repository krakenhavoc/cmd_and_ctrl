package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Elven Bow — Artifact — Equipment {G} (Lord of the Rings Commander /
// Dominaria Remastered):
//
//	"When this Equipment enters, you may pay {2}. If you do, create a 1/1
//	 green Elf Warrior creature token, then attach this Equipment to it.
//	 Equipped creature gets +1/+2 and has reach.
//	 Equip {3}"
//
// The entry trigger is a MayPay whose paid branch is Dwarven
// Shortsword's token-and-attach.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "43daed96-2216-4f17-82ce-3eccd5d69936",
		Name:         "Elven Bow",
		Completeness: effects.CompletenessFull,
		Static:       []game.StaticAbility{effects.PumpAttached(1, 2), effects.GrantToAttached("reach")},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Elven Bow — you may pay {2}; if you do, create a 1/1 green Elf Warrior creature token, then attach this Equipment to it",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{2}",
						Question: "Elven Bow — pay {2} to create a 1/1 green Elf Warrior creature token and attach this Equipment to it?",
						Then:     elvenBowPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{effects.EquipAbility("{3}")},
	})
}

// elvenBowPaid is the paid branch. An on-disk key.
var elvenBowPaid = effects.PayUnlessBranches("pay-unless/elven-bow-token-and-attach",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateTokenAndAttachSource(ctx.Game, ctx.Item, "1/1 green Elf Warrior", true)
	}, nil)
