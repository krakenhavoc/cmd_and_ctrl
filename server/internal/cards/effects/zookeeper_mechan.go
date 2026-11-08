package effects

// Zookeeper Mechan — Artifact Creature — Robot {1}{R}, 1/3 (Edge of
// Eternities, #4461):
//
//	"{T}: Add {R}.
//	 {6}{R}: Target creature you control gets +4/+0 until end of turn.
//	 Activate only as a sorcery."
//
// A creature mana ability (summoning sickness stops the {T}, CR 302.6)
// and the Mercenary token's sorcery-speed pump (boostTargetUntilEOT), read
// at resolution: a target that left or changed controller in response
// gets nothing (CR 608.2b). The Mechan may target itself. The row
// declares the +4/+0 on its target (BenefitUntilEOT) and no
// Purpose.Answers: a sorcery-speed row is never offered while an item
// waits on the stack (ADR 0142).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "6d4d98a8-a5a3-4cc1-9277-c45dafe40ea0",
		Name:         "Zookeeper Mechan",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:     ManaAbilityCost{Tap: true},
			Produced: "{R}",
			Label:    "Add {R}",
		}},
		Activated: []ActivatedAbility{{
			Label:        "{6}{R}: Target creature you control gets +4/+0 until end of turn. Activate only as a sorcery.",
			Purpose:      ForTargets(BenefitUntilEOT(0, 4, 0)),
			Cost:         ManaCost("{6}{R}"),
			Targets:      TargetCreature("target creature you control", YouControl()),
			SorcerySpeed: true,
			Effect:       boostTargetUntilEOT("Zookeeper Mechan — +4/+0", 4, 0),
		}},
	})
}
