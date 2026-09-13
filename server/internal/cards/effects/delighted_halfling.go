package effects

// Delighted Halfling — Creature — Halfling Citizen {G}, 1/2:
//
//	"{T}: Add {C}."
//	"{T}: Add one mana of any color. Spend this mana only to cast a
//	 legendary spell, and that spell can't be countered."
//
// Rank 150, and the highest-ranked card in the restricted-mana group.
// Two abilities: a painless colourless one anybody can use, and a
// colour-fixing one whose mana is locked to legendary spells — which
// in Commander means "your commander", every time.
//
// The restriction is the point of the card, and it is enforced at
// SPEND time, not just stamped at production. `ManaRestrictSupertype`
// puts "supertype:Legendary" on the token; the cast path builds a
// ManaSpendContext from the spell being cast and the pool solver
// refuses the token for anything that is not legendary. Without that
// second half the ability would be a plain "{T}: Add one mana of any
// color" on a one-drop — Birds of Paradise with an extra body — which
// is the #259 direction and the reason both halves landed together in
// #352.
//
// The restriction also means the second ability drops out of auto-tap
// planning (autoTapAbilityFor skips restricted abilities): tapping the
// Halfling for a restricted colour is a decision the planner cannot
// make, because it may waste the only mana that could have cast the
// commander. The colourless ability stays auto-tappable, so a Halfling
// is still a mana source to the planner, just a colourless one. Same
// arrangement the painlands and Talismans have.
//
// Summoning sickness applies to both abilities — ActivateManaAbility
// has enforced CR 302.1 on creature tap abilities since #233.
//
// Simplification, declared: "and that spell can't be countered" is
// INERT. Nothing in the counter path reads a per-spell uncounterable
// marker; that belongs to the protection / prevention family (#95).
// The mana restriction is fully enforced, so the card is strictly
// weaker than printed rather than stronger — the acceptable direction.
func init() {
	Register(Spec{
		OracleID:     "f9d3b046-0b95-4103-a630-4b3fb88bb60b",
		Name:         "Delighted Halfling",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Spells cast with the Halfling's colored mana can still be countered."},
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{W|U|B|R|G}",
				Label:    "Add one mana of any color (legendary spells only)",
				// "one mana of any color" with no mention of the
				// commander's identity — the restriction is the
				// narrowing, not the command zone.
				IgnoreCommanderIdentity: true,
				Restrictions: []string{
					ManaRestrictCast,
					ManaRestrictSupertype("Legendary"),
				},
			},
		},
	})
}
