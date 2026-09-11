package effects

// Phyrexian Tower — Legendary Land (EDHREC rank 198):
//
//	"{T}: Add {C}.
//	 {T}, Sacrifice a creature: Add {B}{B}."
//
// A land that is also a free sacrifice outlet — the aristocrats
// deck's best land, and a ramp spell every turn a creature is worth
// less than two mana. Two separate mana abilities, both real:
//
//   - The colorless half is an ordinary tap ability, and it sits at
//     index 0 so the auto-tapper reaches for it (game.autoTapAbilityFor
//     takes the first tap ability with no sacrifice cost) and never
//     eats a creature on the player's behalf.
//   - The black half is Ashnod's Altar's cost with a tap added:
//     ManaAbilityCost carries both components, the engine validates
//     both before paying either, and the sacrifice fires
//     EventSacrifice plus the dies-triggers before the mana is spent
//     (CR 605.3a — the ability doesn't use the stack, but the
//     sacrifice still triggers).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "1861e642-21d5-4232-89f3-b5557f2946c1",
		Name:     "Phyrexian Tower",
		ManaAbilities: []ManaAbility{
			{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{C}",
				Label:    "Add {C}",
			},
			{
				Cost: ManaAbilityCost{
					Tap:            true,
					SacrificeOther: SacrificeACreature().SacrificeOther,
				},
				Produced: "{B}{B}",
				Label:    "{T}, Sacrifice a creature: Add {B}{B}",
			},
		},
	})
}
