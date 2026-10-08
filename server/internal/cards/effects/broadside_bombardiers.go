package effects

// Broadside Bombardiers — Creature — Goblin Pirate {2}{R}, 2/2:
//
//	"Menace, haste
//	 Boast — Sacrifice another creature or artifact: This creature deals damage
//	 equal to 2 plus the sacrificed permanent's mana value to any target.
//	 (Activate only if this creature attacked this turn and only once each turn.)"
//
// Boast (CR 702.142a) is built with the Boast constructor (boast.go):
// the engine reads the attack record and the activation tally, so the
// card names neither. The activation is spent at the announce, whether
// or not it resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "e0c00a74-bdf8-4fee-8042-42c5878e9e3c",
		Name:            "Broadside Bombardiers",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace", "haste"},
		Activated: []ActivatedAbility{
			BoastTargeting(
				"Sacrifice another creature or artifact: This creature deals damage equal to 2 plus the sacrificed permanent's mana value to any target.",
				SacrificeAnotherN(1, "another creature or artifact", Or(Creature(), Artifact())),
				TargetAny(), sourceDealsComputedDamageToEachLegalTarget(func(ctx *Context) int {
					return 2 + sacrificedManaValue(ctx)
				})),
		},
	})
}
