package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Gaea's Will — Sorcery, no mana cost:
//
//	"Suspend 4—{G}
//	 Until end of turn, you may play lands and cast spells from your
//	 graveyard.
//	 If a card would be put into your graveyard from anywhere this turn,
//	 exile that card instead."
//
// Yawgmoth's Will's clause pair (GraveyardPlayThisTurn) behind Ancestral
// Vision's suspend. It has no mana cost, so suspend is its only way to
// be cast (CR 118.6).
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "87efff06-b6cb-4a8f-937b-e50e367fd896",
		Name:         "Gaea's Will",
		Completeness: CompletenessFull,
		SpecialActions: []game.SpecialAction{
			Suspend(4, "{G}"),
		},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return YawgmothsWillThisTurn().Apply(ctx)
		},
	})
}
