package effects

// Endrider Catalyzer — Creature — Human Warrior {1}{R}, 3/1:
//
//	"Start your engines!
//	 Max speed — {T}: Add {R}{R}."
//
// ADR 0138 (#2122, #2711). A mana ability, so it skips the stack; it is
// available only while its controller has max speed, and a creature's tap
// ability still needs the creature to be free of summoning sickness
// (CR 302.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "b4d970ed-f332-440c-8c2f-8a8a0191410b",
		Name:            "Endrider Catalyzer",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		ManaAbilities: []ManaAbility{
			MaxSpeedMana(ManaAbility{
				Cost:     ManaAbilityCost{Tap: true},
				Produced: "{R}{R}",
				Label:    "Max speed — {T}: Add {R}{R}.",
			}),
		},
	})
}
