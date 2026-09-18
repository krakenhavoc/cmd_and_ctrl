package effects

// Zhur-Taa Druid — Creature — Human Druid {R}{G}, 1/1 (EDHREC rank
// 4194):
//
//	"{T}: Add {G}.
//	 Whenever you tap this creature for mana, it deals 1 damage to
//	 each opponent."
//
// A two-mana dork that pings the whole table every time it ramps you.
// At a four-player game that is three damage a turn for free, which
// is why it shows up in Gruul and in group-slug decks rather than in
// ramp decks proper.
//
// The damage is a RIDER on the mana ability, not a separate trigger,
// and the printed text is what forces that: "whenever you tap this
// creature FOR MANA" fires only for a mana activation. Written as an
// EventTapCard trigger — City of Brass's shape — it would also fire
// when an opponent's Icy Manipulator taps the Druid, when it is
// tapped to crew a Vehicle, and when it attacks, which is a strictly
// better card and the direction this catalog does not ship (#259).
//
// A rider also gets the timing right for free: it runs inside the
// same atomic mana-ability resolution (CR 605.3b), so nobody gets
// priority between the {G} landing in the pool and the three points
// of damage, and the damage happens even at 1 life — a rider is not
// a cost.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5979310e-38c8-489f-ab9e-2723af98a3a5",
		Name:         "Zhur-Taa Druid",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}. This creature deals 1 damage to each opponent",
			Rider:    b40PingEachOpponent(1),
		}},
	})
}
