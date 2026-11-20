package effects

// Bant Sureblade — Creature — Human Soldier {G/U}{W}, 2/1 (Alara Reborn,
// #5586):
//
//	"As long as you control another multicolored permanent, this creature
//	 gets +1/+1 and has first strike."
//
// The blade cycle's shared static (arbB1Blade): a live +1/+1 and a first
// strike grant on the source alone, read through the layers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "610adb2b-9d52-4b70-92d1-0c7adeb93552",
		Name:         "Bant Sureblade",
		Completeness: CompletenessFull,
		Static:       arbB1Blade("first strike"),
	})
}
