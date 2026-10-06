package effects

// Ancient Tomb — Land:
//
//	"{T}: Add {C}{C}. This land deals 2 damage to you."
//
// The purest statement of the rider shape in the format: one
// ability, two colorless mana, two damage, no painless alternative.
// Rank 66 by play rate and, until this batch, a land that produced
// nothing whatsoever — "Land" with no basic supertype gets no
// synthetic ability from the engine.
//
// The two damage is a rider, so:
//
//   - it is not optional and not a cost — the land taps for {C}{C}
//     at any life total above 0, and the 2 damage happens after the
//     mana is already in the pool;
//   - a player at 2 or less who taps it loses, and loses NOW:
//     ActivateManaAbility runs a state-based-action pass on the way
//     out of any activation whose rider fired.
//
// The rider is declared as PainToYou, so the auto-tapper can price it
// (#2392): Ancient Tomb is a pain-tier source, spent only when nothing
// painless can pay, and never when its 2 damage would take its
// controller to 0. Below that it taps by hand from the ability menu.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "23467047-6dba-4498-b783-1ebc4f74b8c2",
		Name:         "Ancient Tomb",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:      ManaAbilityCost{Tap: true},
			Produced:  "{C}{C}",
			Label:     "Add {C}{C}. This land deals 2 damage to you.",
			PainToYou: 2,
		}},
	})
}
