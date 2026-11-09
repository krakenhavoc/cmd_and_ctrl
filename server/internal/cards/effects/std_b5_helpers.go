package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// EntersSacrificeUnlessYouTap is "When this land enters, sacrifice it
// unless you tap an untapped <noun> you control" (Public Thoroughfare,
// Command Bridge): Koskun Falls' tap payment (game.PayActionTap over
// `of`) behind the pay-unless prompt, with the shared "sacrifice this
// permanent" decline. With nothing untapped to tap there is no payment to
// make, so the permanent is sacrificed. A source that left in response, or
// left and came back (CR 400.7), asks nothing.
func EntersSacrificeUnlessYouTap(name, noun, plural string, of game.PermanentQuery) game.TriggeredAbility {
	return WhenThisEnters(name+" — sacrifice it unless you tap an untapped "+noun+" you control",
		func(g *game.Game, item *game.StackItem) error {
			if !onBattlefield(g, item.SourceCardID) || sourceIsNewObject(g, item) {
				return nil
			}
			pay := TapPayment(1, noun, plural, of)
			words, action := pay.prompt()
			return UpkeepPayUnless{
				Chooser:  item.Controller,
				Cost:     words,
				Action:   action,
				Question: name + " — tap an untapped " + noun + " you control or sacrifice " + name + "?",
				Else:     upkeepSacrificeSource,
			}.Apply(NewContext(g, item))
		})
}

// DroneBlockerToken is the Drone's card value for CreateToken.
// DroneBlockerToken is Pinnacle Emissary's Drone: the same printed token as
// Desculpting Blast's (printedDroneToken, slug drone-blocks-only-fliers).
func DroneBlockerToken() game.Card { return tokenFromCatalog(printedDroneToken) }
