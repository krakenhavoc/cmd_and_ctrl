package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Raise the Alarm — Instant {1}{W} (EDHREC rank 4486):
//
//	"Create two 1/1 white Soldier creature tokens."
//
// Dragon Fodder in white, at instant speed. The INSTANT is the whole
// reason the card is played over the sorcery: two blockers appear
// after attackers are declared, and two bodies appear at the end of
// an opponent's turn for a sacrifice deck to spend on your own. The
// timing is the card's type line, which the engine already honours,
// so the Spec is the token clause and nothing else.
//
// The tokens are ordinary white Soldiers — they count for every
// Soldier lord and every "creatures you control" anthem, and they are
// their controller's to sacrifice.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5b2364d7-a811-4595-a1b4-224c70555ffa",
		Name:         "Raise the Alarm",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{
				Controller: ctx.Controller(),
				Template:   TokenCard("1/1 white Soldier"),
				N:          2,
			}.Apply(ctx)
		},
	})
}
