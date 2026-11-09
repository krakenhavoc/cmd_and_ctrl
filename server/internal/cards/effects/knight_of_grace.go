package effects

// Knight of Grace — Creature — Human Knight {1}{W}, 2/2 (Foundations, #4465):
//
//	"First strike
//	 Hexproof from black
//	 This creature gets +1/+0 as long as any player controls a black
//	 permanent."
//
// Hexproof from black is the canonical token (game.HexproofFromColor); the
// pump is a layer 7c self static that reads every battlefield permanent's
// colour through the layered view (registerHexproofKnight).
//
// No simplification.
func init() {
	registerHexproofKnight("7b51a979-47e0-4b98-8ec5-8064c83422e1", "Knight of Grace", "B", "B", "black")
}
