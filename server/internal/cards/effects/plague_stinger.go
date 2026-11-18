package effects

// Plague Stinger — Creature — Phyrexian Insect Horror {1}{B}, 1/1:
//
//	"Flying
//	 Infect (This creature deals damage to creatures in the form of -1/-1
//	 counters and to players in the form of poison counters.)"
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "047e761e-7bbe-403b-86d7-cb9fb0b20b21",
		Name:            "Plague Stinger",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "infect"},
	})
}
