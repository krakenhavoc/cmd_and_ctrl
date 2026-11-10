package effects

import (
	"fmt"
	"slices"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// forage.go — #5184: the card side of the forage keyword action (CR
// 701.61a, game/forage.go). To forage is to exile three cards from your
// graveyard or sacrifice a Food, and Bloomburrow prints it as
//
//	a cost              Plus(ManaCost("{2}"), Forage())       Camellia
//	an additional cost  AdditionalCost: ForageOrPay("{B}")    Feed the Cycle
//	an effect           MayForage(ctx, question, then)        Treetop Sentries
//	a trigger           WheneverYouForage(label, effect)      Corpseberry Cultivator
//
// Each is the owner's own either/or (OrCost, EitherCost, EitherPayment)
// over the plain graveyard-exile and sacrifice payments, marked as a
// forage, so it is payable exactly when one of the two ways is (CR
// 118.3) and the view, the enumerator and the engine read the same walks.

// Forage is "Forage" as part of an activated ability's cost (CR 701.61a,
// #5184): exile three cards from your graveyard, or sacrifice a Food.
//
//	Cost: Plus(ManaCost("{2}"), Forage()), // Camellia, the Seedmiser
//
// It is an OrCost over game.ForageAbilityBranches, marked Forage: the
// activator announces the branch (`cost_branch`: 0 exiles three cards,
// 1 sacrifices a Food) and names the cards or the Food with it, and a
// branch they can't pay can't be chosen, so with fewer than three cards
// in the graveyard and no Food the ability can't be activated. Paying it
// records PaidCost.Foraged and emits game.EventForage. A mana ability
// declares the same cost as ManaAbilityCost.Forage (#5397).
func Forage() game.AbilityCost {
	return game.AbilityCost{Either: game.ForageAbilityBranches(), Forage: true}
}

// ForagePayment is forage as a payment made while something resolves
// (game.ForagePayAction): one either payment whose answer names the way
// (`pay_option` 0 exiles three cards from the graveyard, 1 sacrifices a
// Food), and paying either is the player foraging.
func ForagePayment() UpkeepPayment {
	return UpkeepPayment{Action: game.ForagePayAction(), Words: "Forage (exile three cards from your graveyard or sacrifice a Food)"}
}

// MayForage is "you may forage. If you do, <then>" (#5184): Treetop
// Sentries, Bushy Bodyguard, Corpseberry Cultivator. One prompt to the
// controller naming the way they forage, offered only when they can
// forage at all — with fewer than three cards in their graveyard and no
// Food nobody is asked and "if you do" is false (CR 118.3). `then` runs
// once the forage is paid; "when you do" is a WhenYouDo applied from it.
func MayForage(ctx *Context, question string, then game.PayUnlessElse) error {
	return MayPay{
		Chooser:  ctx.Controller(),
		Pay:      ForagePayment(),
		Question: question,
		Then:     then,
	}.Apply(ctx)
}

// mayForageNothingMore is the paid branch of a "you may forage" with
// nothing printed after it (Corpseberry Cultivator's beginning of
// combat): the forage is the whole point, and "whenever you forage" is
// what pays off.
var mayForageNothingMore = PayUnlessBranches("pay-unless/forage-only",
	func(*Context, game.PayUnlessAnswer) error { return nil }, nil)

// WheneverYouForage is "Whenever you forage, <effect>" (Corpseberry
// Cultivator): one trigger per forage, however it was paid — a cost, an
// additional cost or "you may forage" — off game.EventForage.
func WheneverYouForage(label string, effect Effect) game.TriggeredAbility {
	return On(game.EventForage, ByYou, label, effect)
}

// checkManaForage is Register's guard on a mana ability's forage
// (#5397, ManaAbilityCost.Forage). The branch is folded into the
// ability's sacrifice-another or exile-cards clause
// (game.ChosenManaAbilityShape), so the rest of the cost may carry
// neither, nor a discard; and a forage is paid from the battlefield,
// where the Food and the activator's graveyard are, so a hand ability
// (a Spirit Guide's shape) may not forage. Card-file mistakes, refused
// at boot.
func checkManaForage(name string, i int, ma ManaAbility) {
	if !ma.Cost.Forage {
		return
	}
	if ma.Cost.SacrificeOther != nil || ma.Cost.DiscardCards != nil || ma.Cost.ExileCards != nil {
		panic(fmt.Sprintf("effects.Register: %q mana ability %d forages beside its own sacrifice, discard or exile clause — the forage branch would replace it (game.ChosenManaAbilityShape)", name, i))
	}
	if ma.Cost.ExileSelf || len(ma.Zones) > 0 && !slices.Equal(ma.Zones, []game.ZoneKind{game.ZoneBattlefield}) {
		panic(fmt.Sprintf("effects.Register: %q mana ability %d forages off the battlefield or exiles its source — a forage mana ability is a permanent's", name, i))
	}
}
