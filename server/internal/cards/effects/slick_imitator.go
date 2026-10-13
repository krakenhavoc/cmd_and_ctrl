package effects

// Slick Imitator — Creature — Ooze {1}{U}, 1/3:
//
//	"Start your engines!
//	 Max speed — {1}, Sacrifice this creature: Copy target spell you
//	 control. You may choose new targets for the copy. (A copy of a
//	 permanent spell becomes a token.)"
//
// ADR 0138 (#2711). The target is any spell its controller controls; the
// sacrifice is paid at announce, so the copy resolves without the
// Imitator. Lithoform Engine's copy body, with the CR 707.10c re-target
// offer.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5482ddb1-7f7d-492a-aa3d-67c037062cc3",
		Name:            "Slick Imitator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{StartYourEngines},
		Activated: []ActivatedAbility{
			MaxSpeedActivated(ActivatedAbility{
				Label:   "Max speed — {1}, Sacrifice this creature: Copy target spell you control. You may choose new targets for the copy.",
				Cost:    Plus(ManaCost("{1}"), SacrificeThis()),
				Targets: TargetSpell("target spell you control", YouControl()),
				Effect:  copyTargetedSpell,
			}),
		},
	})
}
