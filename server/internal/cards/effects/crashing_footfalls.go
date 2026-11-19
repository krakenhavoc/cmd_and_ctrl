package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Crashing Footfalls — Sorcery with no mana cost (Special Guests, S82 BNS):
//
//	"Suspend 4—{G}
//	 Create two 4/4 green Rhino creature tokens with trample."
//
// Ancestral Vision's shape: no mana cost means it can never be cast from
// hand (CR 118.6), so suspend is the whole of its playability; the last time
// counter's free cast resolves it. Two Rhinos arrive together.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "2b166d97-c2c1-4642-b151-9bb1ad32e362",
		Name:         "Crashing Footfalls",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Suspend(4, "{G}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return CreateToken{Template: TokenCard("4/4 green Rhino with trample"), N: 2}.Apply(ctx)
		},
	})
}
