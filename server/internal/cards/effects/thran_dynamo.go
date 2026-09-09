package effects

// Thran Dynamo — Artifact {4}:
//
//	"{T}: Add {C}{C}{C}."
//
// Three colourless off one tap is why this is still played in
// colourless-hungry and high-curve decks. Pure mana ability, no
// resolve hook: three fixed slots drop straight into the pool with
// no picker, the same shape as Sol Ring's two.
func init() {
	Register(Spec{
		OracleID: "a699c663-8131-4045-9265-a83e86609374",
		Name:     "Thran Dynamo",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}{C}",
			Label:    "Add {C}{C}{C}",
		}},
	})
}
