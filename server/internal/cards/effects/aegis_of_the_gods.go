package effects

// Aegis of the Gods — Enchantment Creature — Human Soldier {1}{W},
// 2/1:
//
//	"You have hexproof."
//
// Leyline of Sanctity on a body, and the second proof that the
// derived half of #1197 really is derived: the grant lives on the
// battlefield, so killing the Aegis takes the hexproof with it and
// nothing has to remember to undo anything. A "set on enter, restore
// on leave" design would have to answer "restore to what?" with a
// Leyline also out; this one has no stored value to strand.
//
// It is also the card that makes the composition case real rather
// than theoretical — Aegis plus Leyline is two sources of one
// ability, and the engine reads both without either knowing about
// the other.
//
// The creature type line matters here in a way it does not on the
// Leyline: an Aegis that stops being a creature (or loses its
// abilities to a layer-6 effect) stops granting, because the reader
// asks CatalogAbilityKey rather than the card's printed identity.
//
// No simplifications. The whole card is one clause and the clause is
// implemented.
func init() {
	Register(Spec{
		OracleID:       "c5bfc1b9-a55d-4608-a6f7-bb62cb8dc3c6",
		Name:           "Aegis of the Gods",
		Completeness:   CompletenessFull,
		PlayerKeywords: []string{KeywordHexproof},
	})
}
