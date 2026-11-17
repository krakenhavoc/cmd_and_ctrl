package effects

// Island of Wak-Wak — Land (Arabian Nights, #5641):
//
//	"{T}: Target creature with flying has base power 0 until end of
//	 turn."
//
// Singing Tree's layer 7b base power 0 on any creature with flying,
// whoever controls it. The land has no mana ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d427e61d-5b30-4d2a-bad2-2e7f016036ca",
		Name:         "Island of Wak-Wak",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{T}: Target creature with flying has base power 0 until end of turn.",
			Cost:    TapCost(),
			Targets: TargetCreature("target creature with flying", HasKeyword("flying")),
			Effect:  basePowerZeroUntilEOT("Island of Wak-Wak — base power 0 until end of turn"),
		}},
	})
}
