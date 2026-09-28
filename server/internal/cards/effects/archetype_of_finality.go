package effects

// Archetype of Finality — Enchantment Creature — Gorgon {4}{B}{B}, 2/3
// (EDHREC rank 5789):
//
//	"Creatures you control have deathtouch.
//	 Creatures your opponents control lose deathtouch and can't have or
//	 gain deathtouch."
//
// The black Archetype, built on #1651's can't-have strip (ADR 0038's
// amendment of 2026-09-28, B2). Deathtouch is read by the damage path
// off the source's effective abilities, so a stripped deathtouch really
// stops killing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "eb4e9ce9-1917-46b1-b01c-645f5920ef9c",
		Name:         "Archetype of Finality",
		Completeness: CompletenessFull,
		Static:       archetypeStatics("deathtouch"),
	})
}
