package effects

// Deranged Whelp — Creature — Wolf {1}{R}, 2/1 (Eldritch Moon, #5541):
//
//	"Menace (This creature can't be blocked except by two or more creatures.)"
//
// Mek'leth Berserker's keyword, carried as printed data.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "6f246109-8c8b-4909-9888-561bc46ad776",
		Name:            "Deranged Whelp",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"menace"},
	})
}
