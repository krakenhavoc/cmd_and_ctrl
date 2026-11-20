package effects

// Skyhunter Cub — Creature — Cat Knight {2}{W}, 2/2 (Mirrodin, #5606):
//
//	"As long as this creature is equipped, it gets +1/+1 and has flying."
//
// Leonin Den-Guard's static with flying (mrdB5WhileEquippedPlusOne):
// a layer 7c boost and a layer 6 grant on the creature itself, both
// read live from the attachment relation (CR 301.5).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "1e9ebdf3-306a-48e1-8096-6a1d289e108c",
		Name:         "Skyhunter Cub",
		Completeness: CompletenessFull,
		Static:       mrdB5WhileEquippedPlusOne("flying"),
	})
}
