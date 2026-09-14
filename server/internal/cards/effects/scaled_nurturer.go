package effects

// Scaled Nurturer — Creature — Dragon Druid {1}{G}, 0/2 (EDHREC rank
// 3669):
//
//	"{T}: Add {G}. When you spend this mana to cast a Dragon creature
//	 spell, you gain 2 life."
//
// The Dragon deck's mana dork, and a Dragon itself for the lords and
// the Thrasher-style payoffs. The mana half is a plain tapped {G},
// with summoning sickness applying as it does to every creature's
// tap ability (CR 302.6).
//
// SANDBOX SIMPLIFICATION — the life-gain rider is NOT implemented,
// Path of Ancestry's posture exactly. "When you spend THIS MANA to
// cast …" is a delayed trigger keyed to the provenance of one mana
// in the pool: the pool's tokens record which permanent produced
// them, but the spend path discards the tokens it pays with and
// reports nothing about them, so there is no seam to hang the
// trigger on. Wiring it needs the cost engine to hand back the
// tokens it spent (which also unlocks sunburst, converge and the
// "mana from a Treasure" family), and that is a mana-pipeline
// change, not a card change.
//
// The direction is WEAKER than printed: the Nurturer still taps for
// the same green; it just never gains the life. A player who wants
// the 2 life can take it manually — the automation simply does not
// grant it.
func init() {
	Register(Spec{
		OracleID:     "13b96709-0e88-476b-9485-956e682bb818",
		Name:         "Scaled Nurturer",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The 2 life never happens when you spend its mana on a Dragon creature spell — it only taps for {G}."},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{G}",
			Label:    "Add {G}",
		}},
	})
}
