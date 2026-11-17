package effects

// Tarkir Duneshaper // Burnished Dunestomper — {W} Creature — Dog Warrior
// 1/2 // Creature — Phyrexian Dog Warrior 4/3 (March of the Machine):
//
//	Front: "{4}{G/P}: Transform this creature. Activate only as a sorcery.
//	        ({G/P} can be paid with either {G} or 2 life.)"
//	Back:  "Trample"
//
// The front is the Phyrexian transform activation, written inline as
// Concealing Curtains writes its own; {G/P} is paid with {G} or 2 life.
// The back face is trample and nothing else, registered under the same
// oracle ID with "#1".
//
// No simplification.
func init() {
	const oracle = "c554ae68-0201-4958-aa63-8159911acb50"
	Register(Spec{
		OracleID:     oracle,
		Name:         "Tarkir Duneshaper",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:        "{4}{G/P}: Transform this creature. Activate only as a sorcery.",
			Cost:         ManaCost("{4}{G/P}"),
			SorcerySpeed: true,
			Effect:       Do(TransformThis{}),
		}},
	})
	Register(Spec{
		OracleID:        oracle + "#1",
		Name:            "Burnished Dunestomper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"trample"},
	})
}
