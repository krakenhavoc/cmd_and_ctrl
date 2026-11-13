package effects

// Thunderbolt — Instant {1}{R} (Weatherlight, #5629):
//
//	"Choose one —
//	 • Thunderbolt deals 3 damage to target player or planeswalker.
//	 • Thunderbolt deals 4 damage to target creature with flying."
//
// Two bullets, each with its own target clause (CR 700.2). Glorious
// Decay's second bullet is the flying one.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6ffae2fa-511a-45bf-93c4-5d53c85324f5",
		Name:         "Thunderbolt",
		Completeness: CompletenessFull,
		Modes: ChooseOne(
			ModeDoing("Thunderbolt deals 3 damage to target player or planeswalker.",
				targetPlayerOrPlaneswalker(),
				socB20DamageToTheModeTarget(3)),
			ModeDoing("Thunderbolt deals 4 damage to target creature with flying.",
				TargetCreature("target creature with flying", HasKeyword("flying")),
				socB20DamageToTheModeTarget(4)),
		),
	})
}
