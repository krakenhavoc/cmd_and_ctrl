package effects

// Goblin Flectomancer — Creature — Goblin Wizard {U}{R}{R}, 2/2:
//
//	"Sacrifice this creature: You may change the targets of target
//	 instant or sorcery spell."
//
// CR 115.7a with the "you may": every slot of the spell may move,
// declining is allowed, and there is no single-target restriction. The
// sacrifice is the cost, paid at announce, so the Flectomancer is gone
// before the ability resolves and the retarget is not tied to it.
func init() {
	Register(Spec{
		OracleID:     "7c3171da-7a95-4f8e-ba30-c31af21c1828",
		Name:         "Goblin Flectomancer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: You may change the targets of target instant or sorcery spell.",
			Cost:    SacrificeThis(),
			Targets: instantOrSorcerySpell("target instant or sorcery spell"),
			Effect:  chooseNewTargets("Goblin Flectomancer — change the targets"),
		}},
	})
}
