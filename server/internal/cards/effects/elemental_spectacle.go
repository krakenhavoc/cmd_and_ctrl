package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Elemental Spectacle — Sorcery {5}{G} (Lorwyn Eclipsed Commander, #5368):
//
//	"Vivid — Create a number of 5/5 red and green Elemental creature
//	 tokens equal to the number of colors among permanents you control.
//	 Then you gain life equal to the number of creatures you control."
//
// The colour count is read as the spell resolves, before the tokens exist
// (the tokens are red and green and so may add colours only for the NEXT
// reader); the life gained then counts the creatures you control after the
// tokens have entered ("Then"), read live.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "dd963396-3f81-49ad-a68e-87b9764d7ac5",
		Name:         "Elemental Spectacle",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Tokens: 1, CreatureTokens: 1},
		OnResolve:    cmdr2S12ElementalSpectacle,
	})
}
