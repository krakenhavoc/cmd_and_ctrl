package effects

// Jegantha, the Wellspring — Legendary Creature — Elemental Elk {4}{R/G}, 5/5:
//
//	"Companion — No card in your starting deck has more than one of the
//	 same mana symbol in its mana cost. (If this card is your chosen
//	 companion, you may put it into your hand from outside the game for
//	 {3} as a sorcery.)
//	 {T}: Add {W}{U}{B}{R}{G}. This mana can't be spent to pay generic
//	 mana costs."
//
// #2170. The restriction is the whole card: five mana for a tap is
// ramp only if it can pay the coloured symbols of a cost, and
// ManaRestrictNoGeneric is how the engine says so. Every payer reads it
// at the one place the generic part of a cost is paid (the pool solver,
// MissingFor, the auto-tap top-up's pool credit, and so the view and
// the bot's affordability probe), so {N}, {X}, a cost increase and
// commander tax are all refused it, while a coloured, hybrid or
// Phyrexian symbol takes it.
//
// Like Delighted Halfling's coloured ability, a restricted ability is
// never planned by the auto-tapper (autoTapAbilityFor skips it), so the
// player taps Jegantha by hand and the mana floats; weaker than printed,
// never stronger.
//
// The companion clause is not modelled. It is a deck-building condition
// and a special action from outside the game, and the engine has no
// companion zone; deck.Resolve rejects a companion named as a commander
// and checks no companion condition anywhere. In the main deck Jegantha
// is simply a creature, which is weaker than printed.
func init() {
	Register(Spec{
		OracleID:     "f0587c55-06c1-4930-a62e-668f37464ce9",
		Name:         "Jegantha, the Wellspring",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Companion isn't supported — Jegantha can't be put into your hand from outside the game, and the deck checker doesn't enforce its deck-building condition."},
		ManaAbilities: []ManaAbility{{
			Cost:         ManaAbilityCost{Tap: true},
			Produced:     "{W}{U}{B}{R}{G}",
			Label:        "Add {W}{U}{B}{R}{G} (can't pay generic costs)",
			Restrictions: []string{ManaRestrictNoGeneric},
		}},
	})
}
