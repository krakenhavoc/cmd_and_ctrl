package effects

// Archetype of Courage — Enchantment Creature — Human Soldier
// {1}{W}{W}, 2/2 (EDHREC rank 4481):
//
//	"Creatures you control have first strike.
//	 Creatures your opponents control lose first strike and can't have
//	 or gain first strike."
//
// The white Archetype. Half of it is an anthem-shaped keyword grant
// and half of it is a keyword LOCKOUT, and the lockout is the half
// that wins games: your blocks never trade and your attacks always do.
//
// It used to ship only the first line. A plain layer-6 removal would
// have missed first strike granted by anything with a newer timestamp
// (CR 613.7), so it looked like it worked and did not. #1651 (ADR
// 0038's amendment of 2026-09-28, B2) strips "can't have" keywords
// after the whole layer-6 bucket, so LoseAndCantHave now beats a grant
// of any timestamp, as the printed "can't" requires (CR 101.2).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "79b48704-480d-4905-b87a-40b127894670",
		Name:         "Archetype of Courage",
		Completeness: CompletenessFull,
		Static:       archetypeStatics("first strike"),
	})
}
