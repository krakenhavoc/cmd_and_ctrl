package effects

// War Room — Land:
//
//	"{T}: Add {C}.
//	 {3}, {T}, Pay life equal to the number of colors in your
//	 commanders' color identity: Draw a card."
//
// A colourless rock with a card-draw tax that scales with the table's
// commander count — a mono-colour deck pays 1 life, a five-colour
// partner pair pays 5.
//
// S49 sandbox simplification: only the mana half is implemented. The
// draw ability's life cost isn't a fixed number — it depends on how
// many colors are in your commanders' color identity — and the
// engine's activated-ability cost only holds a fixed life amount, so
// there's no way to charge the right price for every deck. Shipping
// it at a made-up flat cost would make the card stronger or weaker
// than printed depending on the deck, so the draw ability is left off
// entirely; the land only taps for {C}.
func init() {
	Register(Spec{
		OracleID:     "71c52bf5-2a5d-488e-8b15-7ef290e4b77d",
		Name:         "War Room",
		Completeness: CompletenessCaveats,
		Caveats:      []string{`The "{3}, {T}, Pay life equal to the number of colors in your commanders' color identity: Draw a card" ability isn't implemented — that life cost scales per deck, which the engine can't charge yet. The land only taps for colorless mana.`},
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}",
			Label:    "Add {C}",
		}},
	})
}
