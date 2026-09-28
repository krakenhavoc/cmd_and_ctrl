package effects

// Archetype of Aggression — Enchantment Creature — Human Warrior
// {1}{R}{R}, 3/2 (EDHREC rank 2850):
//
//	"Creatures you control have trample.
//	 Creatures your opponents control lose trample and can't have or
//	 gain trample."
//
// Two layer-6 statics: a keyword grant to the controller's creatures
// (Lord of Atlantis's shape) and LoseAndCantHave over everyone else's.
//
// It used to ship with a declared gap. The second line was a plain
// layer-6 removal, so a trample grant with a NEWER timestamp — an
// opponent's Garruk's Uprising, or a catalog trampler's own
// PrintedKeywords static entering afterwards — applied after it and
// kept its trample. #1651 (ADR 0038's amendment of 2026-09-28, B2)
// strips "can't have" keywords after the whole layer-6 bucket, so the
// CR 101.2 "can't" now beats a grant of any timestamp.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "263408e6-b315-4af5-8cb8-3fd1aa88e48c",
		Name:         "Archetype of Aggression",
		Completeness: CompletenessFull,
		Static:       archetypeStatics("trample"),
	})
}
