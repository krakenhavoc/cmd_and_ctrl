package effects

// Nightshade Dryad — Creature — Dryad {1}{G}, 1/2 (EDHREC rank 2769):
//
//	"Deathtouch
//	 {T}: Add {C}.
//	 {T}: Add one mana of any color."
//
// A two-mana dork that blocks. Two mana abilities, as printed — the
// colourless one is its own entry because a client picking "Add {C}"
// should not be shown a colour picker — and the any-colour one keeps
// its printed width rather than narrowing to the commander's identity
// (the text says "any color"). Summoning sickness applies to both,
// since both tap a creature (CR 302.6); the engine enforces that.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "f8263274-8cb4-4eab-b541-c7655b555830",
		Name:            "Nightshade Dryad",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"deathtouch"},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:                    ManaAbilityCost{Tap: true},
				Produced:                "{W|U|B|R|G}",
				Label:                   "Add one mana of any color",
				IgnoreCommanderIdentity: true,
			},
		},
	})
}
