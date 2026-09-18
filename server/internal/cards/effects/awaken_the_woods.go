package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Awaken the Woods — Sorcery {X}{G}{G} (EDHREC rank 1320):
//
//	"Create X 1/1 green Forest Dryad land creature tokens. (They're
//	 affected by summoning sickness.)"
//
// X Dryad Arbors for X+2 mana. Each token is a LAND with the Forest
// subtype, so the engine gives it "{T}: Add {G}" from its type line
// alone (CR 305.6, #354), and a creature, so it is summoning sick
// the turn it arrives — both halves of the reminder text fall out
// of the token's type line with nothing declared. X is the
// announced value; X = 0 makes nothing.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "3cc79d36-1a24-4395-8ad1-915a65db8a60",
		Name:         "Awaken the Woods",
		XMatters:     true,
		Completeness: CompletenessFull,
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			return CreateToken{Controller: ctx.Controller(), Template: TokenCard("1/1 green Dryad"), N: ctx.X()}.Apply(ctx)
		},
	})
}
