package effects

// Hard-Hitting Question — Sorcery {G} (Murders at Karlov Manor, #4469):
//
//	"Target creature you control deals damage equal to its power to
//	 target creature or planeswalker you don't control."
//
// Bite Down's text at sorcery speed: bitingSpell's two clauses, each
// checked on its own at announce and at resolution (CR 608.2b), with the
// biter's power read as the spell resolves.
//
// No simplification.
func init() {
	Register(bitingSpell("05fdefb4-2be7-4d9a-9ab9-14bbbe72aef7", "Hard-Hitting Question"))
}
