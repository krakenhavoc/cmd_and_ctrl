package effects

// Hailstorm Valkyrie — Snow Creature — Angel Wizard {3}{B}, 2/2
// (Kaldheim, #5986):
//
//	"Flying, trample
//	 {S}{S}: This creature gets +2/+2 until end of turn. ({S} can be paid
//	 with one mana from a snow source.)"
//
// Flying and trample are printed keywords. The pump is the shared
// selfPumpRow; each {S} is paid only by mana a snow source made, of any
// colour (CR 107.4h), which the payment, the auto-tapper and the
// legal-move list all enforce since #5986.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "d42ecab7-f6ce-4e50-97de-7aa1ced64ebe",
		Name:            "Hailstorm Valkyrie",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying", "trample"},
		Activated:       []ActivatedAbility{selfPumpRow("Hailstorm Valkyrie", "{S}{S}", 2, 2)},
	})
}
