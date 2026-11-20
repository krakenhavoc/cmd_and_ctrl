package effects

// Jund Hackblade — Creature — Goblin Berserker {B/G}{R}, 2/1 (Alara
// Reborn, #5586):
//
//	"As long as you control another multicolored permanent, this creature
//	 gets +1/+1 and has haste."
//
// The blade cycle's shared static (arbB1Blade): a live +1/+1 and a haste
// grant on the source alone, read through the layers, so it can attack
// the turn it arrives while another multicolored permanent is there.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e0e2184e-d4e3-4620-89c4-8378ce3a59c3",
		Name:         "Jund Hackblade",
		Completeness: CompletenessFull,
		Static:       arbB1Blade("haste"),
	})
}
