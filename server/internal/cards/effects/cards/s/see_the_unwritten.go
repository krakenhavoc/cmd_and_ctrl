package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// See the Unwritten — Khans of Tarkir (S82 OLD b347, #5554):
//
//	Sorcery {4}{G}{G}: Reveal the top eight cards of your library. You may put a
//	creature card from among them onto the battlefield. Put the rest into your
//	graveyard. Ferocious — If you control a creature with power 4 or greater, you
//	may put two creature cards onto the battlefield instead of one.
//
// Ferocious is read through YouControlPowerFourOrGreater (current power, layered). Read as the spell resolves, before the reveal.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "b0060b32-b856-4ff7-a113-cec2dc8f8dc6",
		Name:         "See the Unwritten",
		Completeness: effects.CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			player := ctx.Controller()
			most := 1
			if effects.FerociousYouControlPowerFourOrGreater(ctx) {
				most = 2
			}
			const label = "See the Unwritten — put a creature card (two with ferocious) from among the top eight onto the battlefield, the rest into your graveyard"
			return effects.PutFromLibraryOntoBattlefield{
				Player:   player,
				Cards:    ctx.Game.RevealTopOfLibraryForEffect(player, ctx.Source(), 8, "See the Unwritten — revealed from the top of the library"),
				Match:    effects.Creature(),
				Max:      most,
				Optional: true,
				Label:    label,
				ThenTail: effects.PutRestIntoGraveyardTail,
			}.Apply(ctx)
		},
	})
}
