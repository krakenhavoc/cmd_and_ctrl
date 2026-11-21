package cardse

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Emperor Mihail II — Legendary Creature — Merfolk Noble {1}{U}{U}, 3/3:
//
//	"You may look at the top card of your library any time.
//	 You may cast Merfolk spells from the top of your library.
//	 Whenever you cast a Merfolk spell, you may pay {1}. If you do,
//	 create a 1/1 blue Merfolk creature token."
//
// The first two lines are the library-top seam (LibraryTopOwner and a
// top-of-library cast permission narrowed to Merfolk, as Conspicuous Snoop
// does for Goblins). The trigger fires for a Merfolk spell cast from any
// zone, the top of the library included; the {1} is asked as it resolves.
func init() {
	effects.Register(effects.Spec{
		OracleID:          "6c901ff4-15c3-468f-9a1b-4c69a7bbc31a",
		Name:              "Emperor Mihail II",
		Completeness:      effects.CompletenessFull,
		LibraryTopVisible: game.LibraryTopOwner,
		CastPermissions: []game.CastPermission{
			effects.PlayFromTopOfYourLibrary(
				game.PermissionFilter{NonLandOnly: true, CreatureType: "Merfolk"},
				"Cast a Merfolk spell from the top of your library (Emperor Mihail II)"),
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverYouCast(effects.HasSubtype("Merfolk"),
				"Emperor Mihail II — you may pay {1} to create a 1/1 blue Merfolk token",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Emperor Mihail II — pay {1} to create a 1/1 blue Merfolk creature token?",
						Then:     cmdr2B05MihailPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var cmdr2B05MihailPaid = effects.PayUnlessBranches("pay-unless/cmdr2-b05-emperor-mihail-merfolk",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		return effects.CreateToken{Controller: ctx.Controller(), Template: effects.TokenCard("1/1 blue Merfolk"), N: 1}.Apply(ctx)
	}, nil)
