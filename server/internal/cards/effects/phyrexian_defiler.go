package effects

// Phyrexian Defiler — Creature — Phyrexian Carrier {2}{B}{B}, 3/3
// (Urza's Legacy, #5623):
//
//	"{T}, Sacrifice this creature: Target creature gets -3/-3 until end
//	 of turn."
//
// A Carrier through the shared ulgB1SacrificeToShrink row at -3/-3: {T} and
// sacrificing it are both costs, so it waits out summoning sickness
// (CR 302.6) and is gone before the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "bed87e57-d410-478f-ae9d-2d7d00831e5e",
		Name:         "Phyrexian Defiler",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			ulgB1SacrificeToShrink("Phyrexian Defiler", 3),
		},
	})
}
