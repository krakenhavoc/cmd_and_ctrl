package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Deceit — Creature — Elemental Incarnation {4}{U/B}{U/B}, 5/5:
//
//	"When this creature enters, if {U}{U} was spent to cast it, return
//	 up to one other target nonland permanent to its owner's hand.
//	 When this creature enters, if {B}{B} was spent to cast it, target
//	 opponent reveals their hand. You choose a nonland card from it.
//	 That player discards that card.
//	 Evoke {U/B}{U/B}"
//
// Two enters triggers, each behind an intervening if (CR 603.4) that
// reads what paid for the spell this permanent was (Gruul Scrapper's
// read, `ManaSpentToCast`). "{U}{U} was spent" is at least two blue
// mana in the total cost, the hybrid symbols or not, and more than two
// does nothing extra (the 2025-11-17 rulings), so it is Count("U") >= 2.
// With {U}{U}{B}{B} spent both trigger; with one of each, neither does.
// The evoke cost counts too, so an evoke paid {B}{B} still takes a card
// before the CR 702.74a sacrifice.
//
// The second trigger is Thoughtseize's pick (ADR 0116), chosen by the
// trigger's controller (CR 113.8). "Other" on the first is object
// identity (Another).
//
// Declared weaker than printed: with strict mana off the game does not
// record which mana paid, and neither trigger fires.
func init() {
	Register(Spec{
		OracleID:     "d7dddeac-c6ed-4111-ba40-e2830c08d480",
		Name:         "Deceit",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"With strict mana off, the game doesn't see which mana you spent, so neither of Deceit's enters abilities happens."},
		AlternativeCosts: []game.AlternativeCost{
			Evoke("{U/B}{U/B}"),
		},
		Triggered: []game.TriggeredAbility{
			Targeting(
				On(game.EventETB, AllOf(Self, deceitSpent("U")),
					"Deceit — return up to one other nonland permanent to its owner's hand",
					func(g *game.Game, item *game.StackItem) error {
						return bounceTheTarget(item, NewContext(g, item))
					}),
				Another(TargetPermanent("up to one other target nonland permanent", Nonland()).WithCount(0, 1))),
			Targeting(
				On(game.EventETB, AllOf(Self, deceitSpent("B")),
					"Deceit — target opponent reveals their hand, you choose a nonland card",
					TargetRevealsYouChooseDiscardAbility(Nonland(), "nonland card")),
				TargetPlayer("target opponent", Opponent())),
		},
	})
}

// deceitSpent is "if {C}{C} was spent to cast it" for the colour
// symbol c: at least two mana of that colour paid for the spell this
// permanent was.
func deceitSpent(c string) When {
	return func(_ game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return source.ManaSpentToCast().Count(c) >= 2
	}
}
