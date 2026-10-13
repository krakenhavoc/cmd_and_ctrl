package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Replenish — Sorcery {3}{W} (EDHREC rank 3794):
//
//	"Return all enchantment cards from your graveyard to the
//	 battlefield. (Auras with nothing to enchant remain in your
//	 graveyard.)"
//
// The enchantress deck's mass reanimation. Every enchantment
// card in the caster's graveyard comes back under its owner's
// control — an enchantment creature is an enchantment and counts —
// all of them as one entry (#1867), so its own enters-tapped clause
// runs and every ETB trigger sees the whole batch (CR 603.6a). Not targeted, as
// printed, and castable with an empty graveyard to no effect.
//
// Every Aura comes back too (#3158): as each one enters, its controller
// chooses what it enchants (CR 303.4f), one question per Aura, among the
// objects and players already there — not another permanent Replenish
// is returning. An Aura with nothing to enchant stays in the graveyard,
// which is the printed parenthetical (CR 303.4g).
func init() {
	Register(Spec{
		OracleID:     "523ae937-5535-490d-96ea-07f331b5e5ad",
		Name:         "Replenish",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b36ReturnAllEnchantmentCards(ctx)
		},
	})
}
