package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Haunted Library — Enchantment {1}{W} (Innistrad: Crimson Vow Commander):
//
//	"Whenever a creature an opponent controls dies, you may pay {1}. If you
//	 do, create a 1/1 white Spirit creature token with flying."
//
// The condition is Moonstone Eulogist's (the creature left the battlefield
// for a graveyard under an opponent's control, last-known, CR 603.10a). The
// payment is made as the ability resolves (CR 603.5), and the token is made
// only when it was paid.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e787ec12-ba0f-4a85-af6b-422423c90155",
		Name:         "Haunted Library",
		Completeness: CompletenessFull,
		Triggered: []game.TriggeredAbility{
			On(game.EventLTB, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return b18OpponentsCreatureDied(ev, source, g)
			}, "Haunted Library — you may pay {1} to create a 1/1 white Spirit with flying",
				func(g *game.Game, item *game.StackItem) error {
					return MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}",
						Question: "Haunted Library — pay {1} to create a 1/1 white Spirit creature token with flying?",
						Then:     hauntedLibraryPaid,
					}.Apply(NewContext(g, item))
				}),
		},
	})
}

var hauntedLibraryPaid = PayUnlessBranches("pay-unless/haunted-library-spirit",
	func(ctx *Context, _ game.PayUnlessAnswer) error {
		return vowB12SpiritToken().Apply(ctx)
	}, nil)
