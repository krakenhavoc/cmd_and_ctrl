package effects

// Knight of Malice — Creature — Human Knight {1}{B}, 2/2 (Foundations, #4465):
//
//	"First strike
//	 Hexproof from white
//	 This creature gets +1/+0 as long as any player controls a white
//	 permanent."
//
// The mirror of Knight of Grace (registerHexproofKnight).
//
// No simplification.
func init() {
	registerHexproofKnight("b89d912a-65a0-4bc9-9e1d-ff8317458a72", "Knight of Malice", "W", "W", "white")
}
