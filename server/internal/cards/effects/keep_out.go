package effects

// Keep Out — Instant {1}{W} (Lorwyn Eclipsed, #4458):
//
//	"Choose one —
//	 • Keep Out deals 4 damage to target tapped creature.
//	 • Destroy target enchantment."
//
// Two bullets, each with its own target clause, judged as the spell
// resolves (CR 608.2b): a creature untapped in response is an illegal
// target and the damage does nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "61cb927a-4cf8-49d6-ab4b-9e882bc08774",
		Name:         "Keep Out",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Keep Out deals 4 damage to target tapped creature.",
				TargetCreature("target tapped creature", tappedPermanent()),
				DealFixedDamageToModesTarget(4)),
			ModeDoing("Destroy target enchantment.",
				TargetPermanent("target enchantment", Enchantment()),
				DestroyTheModesTarget),
		),
	})
}
