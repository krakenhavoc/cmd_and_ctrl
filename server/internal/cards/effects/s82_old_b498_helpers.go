package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// s82_old_b498_helpers.go — Living Tsunami (Zendikar, #5583): "At the
// beginning of your upkeep, sacrifice this creature unless you return a
// land you control to its owner's hand." It is the Karoo/Bull Elephant
// payment (returnLandsOrSacrifice) with one land of any kind, asked as
// the upkeep trigger resolves.
var s82OldB498ReturnALand = NewReturnLandsOrSacrifice(
	"Living Tsunami — return a land you control to its owner's hand (or sacrifice Living Tsunami)",
	1,
	"choose-cards/s82-old-b498-living-tsunami-return",
	"s82-old-b498/living-tsunami-after-return",
	func(_ *game.Card, _ game.Card) bool { return true },
)

// S82OldB498SacrificeUnlessYouReturnALandAtUpkeep is that trigger.
func S82OldB498SacrificeUnlessYouReturnALandAtUpkeep(name string) game.TriggeredAbility {
	return AtYourUpkeep(name+" — sacrifice it unless you return a land you control to its owner's hand", s82OldB498ReturnALand.enters)
}
