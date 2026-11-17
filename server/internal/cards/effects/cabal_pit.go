package effects

// Cabal Pit — Land (Odyssey, #5613):
//
//	"{T}: Add {B}. This land deals 1 damage to you.
//	 Threshold — {B}, {T}, Sacrifice this land: Target creature gets -2/-2
//	 until end of turn. Activate only if there are seven or more cards in
//	 your graveyard."
//
// Odyssey's black threshold land, built by odyThresholdLand. The -2/-2
// is a layer 7c change (CR 613.4c) that ends at cleanup (CR 514.2); a
// target gone by resolution gets nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(odyThresholdLand("92392467-a22f-4dd7-a0eb-393bef956dc0", "Cabal Pit", "{B}",
		"Target creature gets -2/-2 until end of turn.", ActivatedAbility{
			Targets: TargetCreature("target creature"),
			Purpose: ForTargets(HarmUntilEOT(0, -2, -2)),
			Effect:  pumpAndGrantFirstTarget(-2, -2, "Cabal Pit — -2/-2 until end of turn"),
		}))
}
