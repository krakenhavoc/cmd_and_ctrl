package effects

// Centaur Garden — Land (Odyssey, #5613):
//
//	"{T}: Add {G}. This land deals 1 damage to you.
//	 Threshold — {G}, {T}, Sacrifice this land: Target creature gets +3/+3
//	 until end of turn. Activate only if there are seven or more cards in
//	 your graveyard."
//
// Odyssey's green threshold land, built by odyThresholdLand. The +3/+3
// is a layer 7c change (CR 613.4c) that ends at cleanup (CR 514.2); a
// target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(odyThresholdLand("5cf92fd4-7c0b-4d8e-92f1-53dc2e0476fc", "Centaur Garden", "{G}",
		"Target creature gets +3/+3 until end of turn.", ActivatedAbility{
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(BenefitUntilEOT(0, 3, 3)),
			Effect:  pumpAndGrantFirstTarget(3, 3, "Centaur Garden — +3/+3 until end of turn"),
		}))
}
