package effects

// Grixis Grimblade — Creature — Zombie Warrior {U/R}{B}, 2/1 (Alara
// Reborn, #5586):
//
//	"As long as you control another multicolored permanent, this creature
//	 gets +1/+1 and has deathtouch. (Any amount of damage it deals to a
//	 creature is enough to destroy that creature.)"
//
// The blade cycle's shared static (arbB1Blade): a live +1/+1 and a
// deathtouch grant on the source alone, read through the layers.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "04b1280a-b084-4f65-a3f4-2ba41d6c26fa",
		Name:         "Grixis Grimblade",
		Completeness: CompletenessFull,
		Static:       arbB1Blade("deathtouch"),
	})
}
