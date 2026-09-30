package effects

// Hooting Mandrills — Creature — Ape 4/4 {5}{G}:
//
//	"Delve. Trample"
//
// Delve (CR 702.66, ADR 0100) is `Delve: true` and nothing else: the
// engine prices it, validates the graveyard cards named on
// `delve_ids` and exiles them at CR 601.2h with the spell on the
// stack.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "70e35385-6129-4bd1-861c-df04469566b7",
		Name:            "Hooting Mandrills",
		Completeness:    CompletenessFull,
		Delve:           true,
		PrintedKeywords: []string{"trample"},
	})
}
