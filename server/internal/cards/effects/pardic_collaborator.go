package effects

// Pardic Collaborator — Creature — Human Barbarian {3}{R}, 2/2 (Torment,
// #5612):
//
//	"First strike
//	 {B}: This creature gets +1/+1 until end of turn."
//
// First strike is a printed keyword (CR 702.7b). The pump is the shared
// selfPumpRow Nantuko Shade uses: no tap, so any number of activations,
// each ending at cleanup (CR 514.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "9b06cfae-1655-4b0e-a160-9b32f27c7c9a",
		Name:            "Pardic Collaborator",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"first strike"},
		Activated:       []ActivatedAbility{selfPumpRow("Pardic Collaborator", "{B}", 1, 1)},
	})
}
