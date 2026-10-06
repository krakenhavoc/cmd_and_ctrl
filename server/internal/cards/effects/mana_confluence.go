package effects

// Mana Confluence — Land:
//
//	"{T}, Pay 1 life: Add one mana of any color."
//
// The card that motivated ManaAbilityCost.Life. Everything else in
// this batch loses life as a RIDER — after the mana, unavoidably.
// Mana Confluence loses it as a COST, and the difference is real and
// testable:
//
//   - a cost is validated before anything is paid, so an activation
//     at 0 life is rejected and the land does NOT tap (CR 119.4: you
//     can't pay more life than you have; paying down to exactly 0 is
//     legal);
//   - Ancient Tomb, by contrast, taps happily at 1 life and then
//     kills its controller.
//
// "One mana of any color" is a five-colour pipe with no
// commander-identity narrowing. The narrowing exists only for Arcane
// Signet and Command Tower and their kin, whose printed text names
// the commander's identity; this card says "any color" flatly, and a
// Rakdos deck's Mana Confluence really can produce {G} (offered after
// {B} and {R}). Narrowing it would be a lie about what the card does:
// the player would see a two-colour picker on a card that prints five.
//
// Planned by the auto-tapper since #2392, in its pain tier: only when
// no painless source can pay, and never for life that would take its
// controller to 0 (game.painBudgetFor). The automatic payment pays the
// 1 life as the cost it is. Below that it taps by hand, where CR 119.4
// still allows paying down to exactly 0.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d0ee5bdc-2b69-4b73-9a20-ffcc18783b29",
		Name:         "Mana Confluence",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true, Life: 1},
			Produced: "{W|U|B|R|G}",
			Label:    "Pay 1 life: Add one mana of any color",
		}},
	})
}
