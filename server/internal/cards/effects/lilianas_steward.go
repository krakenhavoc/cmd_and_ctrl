package effects

// Liliana's Steward — Creature — Zombie {B}, 1/2 (Jumpstart, #4892):
//
//	"{T}, Sacrifice this creature: Target opponent discards a card.
//	 Activate only as a sorcery."
//
// The opponent chooses the card they discard (CR 701.9b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "cc1203cb-6758-4454-b78b-341930ff4da1",
		Name:         "Liliana's Steward",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{T}, Sacrifice this creature: Target opponent discards a card. Activate only as a sorcery.",
			Cost:         Plus(TapCost(), SacrificeThis()),
			SorcerySpeed: true,
			Targets:      TargetPlayer("target opponent", Opponent()),
			Effect:       targetOpponentDiscards(1),
		}},
	})
}
