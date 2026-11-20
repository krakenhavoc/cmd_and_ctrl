package effects

// Esper Stormblade — Artifact Creature — Vedalken Wizard {W/B}{U}, 2/1
// (Alara Reborn, #5586):
//
//	"As long as you control another multicolored permanent, this creature
//	 gets +1/+1 and has flying."
//
// The blade cycle's shared static (arbB1Blade): a live +1/+1 and a flying
// grant on the source alone, read through the layers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "186b2fbd-8a59-4447-bd54-45a7c3ff5af2",
		Name:         "Esper Stormblade",
		Completeness: CompletenessFull,
		Static:       arbB1Blade("flying"),
	})
}
