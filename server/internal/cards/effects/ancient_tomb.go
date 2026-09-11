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
// Ancient Tomb is deliberately excluded from the auto-tapper (see
// game.autoTapAbilityFor): its only ability carries the rider, and an
// auto-tap that quietly took 2 off a player's life to save them a
// click would be the wrong trade. It taps by hand from the ability
// menu, which is also how it plays in paper — Ancient Tomb is a
// decision, not a land drop.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID: "23467047-6dba-4498-b783-1ebc4f74b8c2",
		Name:     "Ancient Tomb",
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{C}{C}",
			Label:    "Add {C}{C}. This land deals 2 damage to you.",
			Rider:    PainRider(2),
		}},
	})
}
