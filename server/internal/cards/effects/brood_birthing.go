package effects

// Brood Birthing — Sorcery {1}{R} (Rise of the Eldrazi, #5581):
//
//	"If you control an Eldrazi Spawn, create three 0/1 colorless Eldrazi
//	 Spawn creature tokens. They have "Sacrifice this token: Add {C}."
//	 Otherwise, create one of those tokens."
//
// Whether you control an Eldrazi Spawn is read as the spell resolves: a
// permanent that is both an Eldrazi and a Spawn, token or not.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "e7d1c762-69d0-401d-8035-6e9744baf9d8",
		Name:         "Brood Birthing",
		Completeness: CompletenessFull,
		OnResolve: ResolveSteps(Instruction(func(ctx *Context) error {
			n := 1
			if roeB1YouControlAnEldraziSpawn(ctx) {
				n = 3
			}
			return roeB1Spawns(n).Apply(ctx)
		})),
	})
}
