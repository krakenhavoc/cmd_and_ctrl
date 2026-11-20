package effects

// Naya Hushblade — Creature — Elf Rogue {R/W}{G}, 2/1 (Alara Reborn,
// #5586):
//
//	"As long as you control another multicolored permanent, this creature
//	 gets +1/+1 and has shroud. (It can't be the target of spells or
//	 abilities.)"
//
// The blade cycle's shared static (arbB1Blade): a live +1/+1 and a shroud
// grant on the source alone, read through the layers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d1f986dc-e0f0-4cc6-9dc4-fe97608da332",
		Name:         "Naya Hushblade",
		Completeness: CompletenessFull,
		Static:       arbB1Blade("shroud"),
	})
}
