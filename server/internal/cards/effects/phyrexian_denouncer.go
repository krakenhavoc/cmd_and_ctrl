package effects

// Phyrexian Denouncer — Creature — Phyrexian Carrier {1}{B}, 1/1 (Urza's
// Legacy, #5623):
//
//	"{T}, Sacrifice this creature: Target creature gets -1/-1 until end
//	 of turn."
//
// A Carrier through the shared ulgB1SacrificeToShrink row at -1/-1: {T} and
// sacrificing it are both costs, so it waits out summoning sickness
// (CR 302.6) and is gone before the ability resolves.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b331c5fa-cd97-4fd8-b1df-d2c9e6a5a854",
		Name:         "Phyrexian Denouncer",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			ulgB1SacrificeToShrink("Phyrexian Denouncer", 1),
		},
	})
}
