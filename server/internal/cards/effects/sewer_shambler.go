package effects

// Sewer Shambler — Creature — Zombie {2}{B}, 2/1 (Return to Ravnica, #5566):
//
//	"Swampwalk (This creature can't be blocked as long as defending
//	 player controls a Swamp.) / Scavenge {2}{B} ({2}{B}, Exile this
//	 card from your graveyard: Put a number of +1/+1 counters equal to
//	 this card's power on target creature. Scavenge only as a
//	 sorcery.)"
//
// The printed keywords ride PrintedKeywords; the engine enforces them.
// Scavenge is the shared graveyard ability (Deadbridge Goliath,
// scavenge.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "5d111adc-a91a-437d-8aac-3b246e980bd8",
		Name:            "Sewer Shambler",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"swampwalk"},
		Activated:       []ActivatedAbility{Scavenge("{2}{B}")},
	})
}
