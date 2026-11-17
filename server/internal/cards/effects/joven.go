package effects

// Joven — Legendary Creature — Human Rogue {3}{R}{R}, 3/3 (Homelands, #5635):
//
//	"{R}{R}{R}, {T}: Destroy target noncreature artifact."
//
// The {T} keeps him off the turn he arrives (CR 602.5a). The target is
// an artifact that is not a creature, judged at announce and again at
// resolution (CR 608.2b), so an artifact animated in response is no
// longer a legal target.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "9e767d44-feff-449a-bed9-0866c7bce846",
		Name:         "Joven",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}{R}{R}, {T}: Destroy target noncreature artifact.",
			Cost:    Plus(ManaCost("{R}{R}{R}"), TapCost()),
			Targets: TargetPermanent("target noncreature artifact", Artifact(), Noncreature()),
			Effect:  destroyFirstLegalTarget,
		}},
	})
}
