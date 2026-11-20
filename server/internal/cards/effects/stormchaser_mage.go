package effects

// Stormchaser Mage — Creature — Human Wizard {U}{R}, 1/3 (Oath of the Gatewatch, #5545):
//
//	"Flying, haste
//	 Prowess (Whenever you cast a noncreature spell, this creature gets
//	 +1/+1 until end of turn.)"
//
// Its whole text is printed keywords; prowess is the engine keyword
// Bloodfire Expert uses.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "32902854-2c64-431e-9d9c-9a23e52afb1b",
		Name:            "Stormchaser Mage",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "haste", "prowess"},
	})
}
