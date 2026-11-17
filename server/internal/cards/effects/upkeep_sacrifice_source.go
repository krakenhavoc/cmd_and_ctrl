package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// upkeepSacrificeSource is the "or else" of "sacrifice <this permanent>
// unless you pay ...", keyed (#4264): the permanent is the prompt's
// source, so a table waiting on the payer is a restore point. Stasis,
// Haazda Shield Mate, Koskun Falls and every cumulative upkeep that is
// not a coin flip share it.
var upkeepSacrificeSource = PayUnlessElse("pay-unless/upkeep-sacrifice-source",
	func(ctx *Context, r game.PayUnlessAnswer) error {
		return SacrificePermanent{Target: r.Source}.Apply(ctx)
	})

// UpkeepSacrificeUnlessYouPay is "At the beginning of your upkeep,
// sacrifice <name> unless you pay <cost>." with a mana cost: the
// controller is asked to pay as the trigger resolves, and declining, or
// a "yes" they cannot fund, sacrifices the permanent through the shared
// keyed decline above. The Legends Elder Dragons (Palladia-Mors,
// Vaevictis Asmadi, Nicol Bolas) share it; Haazda Shield Mate and Krosan
// Cloudscraper are the same shape written out.
func UpkeepSacrificeUnlessYouPay(name, cost string) game.TriggeredAbility {
	return AtYourUpkeep(name+" — sacrifice unless you pay "+cost, func(g *game.Game, item *game.StackItem) error {
		return UpkeepPayUnless{
			Chooser:  item.Controller,
			Cost:     cost,
			Question: name + " — pay " + cost + " or sacrifice " + name + "?",
			Else:     upkeepSacrificeSource,
		}.Apply(NewContext(g, item))
	})
}
