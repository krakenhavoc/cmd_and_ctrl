package effects

// Blinding Souleater — Artifact Creature — Phyrexian Cleric {3}, 1/3 (New Phyrexia, #5575):
//
//	"{W/P}, {T}: Tap target creature. ({W/P} can be paid with either {W} or 2 life.)"
//
// Arashin Sunshield's tapper (tapChosenPermanent) for {W/P}; the
// Phyrexian symbol is paid with {W} or 2 life (CR 107.4f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "f5f330a5-14a0-4bfe-ba46-4616692b7bd3",
		Name:         "Blinding Souleater",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{W/P}, {T}: Tap target creature.",
			Cost:    Plus(ManaCost("{W/P}"), TapCost()),
			Targets: TargetCreature("target creature"),
			Effect:  tapChosenPermanent,
		}},
	})
}
