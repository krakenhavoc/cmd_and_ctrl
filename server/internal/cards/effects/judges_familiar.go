package effects

// Judge's Familiar — Creature — Bird {W/U}, 1/1 (Ravnica Remastered, #5948):
//
//	"Flying
//	 Sacrifice this creature: Counter target instant or sorcery spell unless
//	 its controller pays {1}."
//
// The sacrifice is the cost, paid as the ability is activated (CR 602.2b), so
// the Familiar is gone before the spell's controller is asked; Complicate's
// counterTheTargetSpellUnlessPaid asks them as it resolves and counters a spell
// whose controller can't or won't pay (CR 118.12a). A target that left the stack
// leaves the ability with nothing to do (CR 608.2b).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "22b7975c-37a5-4895-b209-bdeeec0c6aea",
		Name:            "Judge's Familiar",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated: []ActivatedAbility{{
			Label:   "Sacrifice this creature: Counter target instant or sorcery spell unless its controller pays {1}.",
			Cost:    SacrificeThis(),
			Targets: instantOrSorcerySpell("target instant or sorcery spell"),
			Effect:  counterTheTargetSpellUnlessPaid("Judge's Familiar", "{1}"),
		}},
	})
}
