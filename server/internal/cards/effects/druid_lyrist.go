package effects

// Druid Lyrist — Creature — Human Druid {G}, 1/1 (Odyssey, #5613):
//
//	"{G}, {T}, Sacrifice this creature: Destroy target enchantment."
//
// Elf Replica's row with a tap symbol added, so a summoning-sick Lyrist
// can't use it (CR 302.6). The sacrifice is part of the cost; a target
// gone in response leaves the ability to do nothing (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b32a10ba-528c-4dc1-9828-b33f5d5a3091",
		Name:         "Druid Lyrist",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{G}, {T}, Sacrifice this creature: Destroy target enchantment.",
			Cost:    Plus(ManaCost("{G}"), TapCost(), SacrificeThis()),
			Targets: TargetPermanent("target enchantment", Enchantment()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
