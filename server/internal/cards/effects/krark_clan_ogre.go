package effects

// Krark-Clan Ogre — Creature — Ogre {3}{R}{R}, 3/3 (Fifth Dawn, #5604):
//
//	"{R}, Sacrifice an artifact: Target creature can't block this turn."
//
// Arena Athlete's "can't block this turn" on any target creature, paid with
// {R} and an artifact.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5de63529-7dea-4018-bf46-afb4bc8b69fd",
		Name:         "Krark-Clan Ogre",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{R}, Sacrifice an artifact: Target creature can't block this turn.",
			Cost:    Plus(ManaCost("{R}"), SacrificeN(1, "an artifact", Artifact())),
			Targets: TargetCreature("target creature"),
			Effect:  targetCreatureCantBlockThisTurn("Krark-Clan Ogre — target creature can't block this turn"),
		}},
	})
}
