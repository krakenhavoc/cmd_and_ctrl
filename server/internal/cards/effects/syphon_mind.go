package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Syphon Mind — Sorcery {3}{B} (EDHREC rank 1049):
//
//	"Each other player discards a card. You draw a card for each card
//	 discarded this way."
//
// The multiplayer Sign in Blood: three cards for four mana at a full
// table. Each opponent's discard is their own choice, through the
// pending-discard modal The Eldest Reborn and Mind Rot use.
//
// Sandbox simplification, declared: the draw count is decided as the
// spell resolves rather than after the discards happen. It is the
// same number either way — every opponent holding at least one card
// discards exactly one, and one with an empty hand discards nothing
// — so the caster draws exactly what the printed card gives; what
// differs is only that the cards are drawn before the opponents have
// picked which card to pitch, because the discard prompt has no
// continuation to hang the draw on. Never stronger than printed.
func init() {
	Register(Spec{
		OracleID:     "abc37d6c-6300-47b5-a679-9db5b83eb54f",
		Name:         "Syphon Mind",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"You draw as the spell resolves, before each opponent has chosen which card to discard — the number of cards is the same."},
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			draws := b09OpponentsHoldingCards(ctx)
			for _, opp := range ctx.Opponents() {
				ctx.Game.DiscardChoiceForEffect(opp, 1)
			}
			return DrawCards{Player: item.Controller, N: draws}.Apply(ctx)
		},
	})
}
