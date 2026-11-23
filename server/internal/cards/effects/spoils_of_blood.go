package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Spoils of Blood — Instant {B} (Commander 2014, #5553):
//
//	"Create an X/X black Horror creature token, where X is the number of
//	 creatures that died this turn."
//
// X is the table-wide tally of creatures that died this turn
// (b11CreaturesDiedThisTurn, Bone Devourer's count), anyone's, read as the
// spell resolves. With no deaths the token is still created, as a 0/0 that
// the state-based actions then put into the graveyard (CR 704.5f).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "26516566-c837-428d-bc81-582c74da2c28",
		Name:         "Spoils of Blood",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: C14BlackHorror(B11CreaturesDiedThisTurn(ctx.Game)), N: 1}.Apply(ctx)
		},
	})
}

// c14BlackHorror is Commander 2014's "X/X black Horror creature token"
// (Spoils of Blood, Flesh Carver): the 1/1 Horror template sized to x, with
// its printed size known so a 0/0 is a 0/0 and not an unknown one.
func C14BlackHorror(x int) game.Card {
	horror := TokenCard("1/1 black Horror")
	horror.Power, horror.Toughness, horror.PrintedPTKnown = x, x, true
	return horror
}
