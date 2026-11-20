package effects

// Crypt Ripper — Creature — Shade {2}{B}{B}, 2/2 (Zendikar, #5583):
//
//	"Haste
//	 {B}: This creature gets +1/+1 until end of turn."
//
// Haste is the printed keyword; the pump is the shared firebreathing
// row (self_pump_row.go).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "18bcf4c5-7abc-4ca5-b805-4bbebf4e433c",
		Name:            "Crypt Ripper",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"haste"},
		Activated:       []ActivatedAbility{selfPumpRow("Crypt Ripper", "{B}", 1, 1)},
	})
}
