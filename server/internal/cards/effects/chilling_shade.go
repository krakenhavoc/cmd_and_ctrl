package effects

// Chilling Shade — Snow Creature — Shade {2}{B}, 1/1 (Coldsnap, #5986):
//
//	"Flying
//	 {S}: This creature gets +1/+1 until end of turn. ({S} can be paid
//	 with one mana from a snow source.)"
//
// Nantuko Shade's selfPumpRow with a snow symbol for its {B}: no tap,
// so any number of activations, each paid by one mana of any colour
// from a snow source (CR 107.4h, #5986) and each ending at cleanup.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:        "1bbcfbad-c4c5-4818-b73c-26d83f7fa22c",
		Name:            "Chilling Shade",
		Completeness:    CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Activated:       []ActivatedAbility{selfPumpRow("Chilling Shade", "{S}", 1, 1)},
	})
}
