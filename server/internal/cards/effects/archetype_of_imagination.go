package effects

// Archetype of Imagination — Enchantment Creature — Human Wizard
// {4}{U}{U}, 3/2 (EDHREC rank 1840):
//
//	"Creatures you control have flying.
//	 Creatures your opponents control lose flying and can't have or
//	 gain flying."
//
// The blue Archetype, built on #1651's can't-have strip (ADR 0038's
// amendment of 2026-09-28, B2). Flying is read by the block check, so
// an opponent's creatures cannot block your fliers unless they have
// reach, and a flying grant cast afterwards gives them nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "a5458de0-0f61-49a3-a013-d90f92559809",
		Name:         "Archetype of Imagination",
		Completeness: CompletenessFull,
		Static:       archetypeStatics("flying"),
	})
}
