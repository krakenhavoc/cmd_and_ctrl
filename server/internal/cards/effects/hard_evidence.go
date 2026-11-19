package effects

// Hard Evidence — Sorcery {U} (The List):
//
//	"Create a 0/3 blue Crab creature token.
//	 Investigate. (Create a Clue token. It's an artifact with "{2},
//	 Sacrifice this token: Draw a card.")"
//
// The Crab first, then the Clue, in the printed order (Sophia, Dogged
// Detective's token-then-investigate shape).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "0728b195-2dfe-4d73-91ae-e658a97b0b9a",
		Name:         "Hard Evidence",
		Completeness: CompletenessFull,
		OnResolve:    ResolveSteps(Instruction(CreateToken{Template: TokenCard("0/3 blue Crab"), N: 1}.Apply), Instruction(Investigate{N: 1}.Apply)),
	})
}
