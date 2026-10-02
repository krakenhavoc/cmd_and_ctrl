package effects

// Flaying Tendrils — Sorcery {1}{B}{B}, devoid:
//
//	"Devoid (This card has no color.)
//	 All creatures get -2/-2 until end of turn. If a creature would die
//	 this turn, exile it instead."
//
// Devoid is colour data (the dump's colour list is empty). "If a
// creature would die this turn" changes no characteristic, so its set is
// read as each creature would die (CR 611.2c, ADR 0108 §1): a creature
// cast after Flaying Tendrils resolved is exiled too if it dies this
// turn. The -2/-2 locks its set as it begins.
//
// No simplifications.
func init() {
	Register(Spec{
		OracleID:     "cc2d016a-af44-427b-a25a-593274369449",
		Name:         "Flaying Tendrils",
		Completeness: CompletenessFull,
		OnResolve:    allCreaturesShrinkThenExileIfTheyDie(-2, false),
	})
}
