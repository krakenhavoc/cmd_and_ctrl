package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brilliant Restoration — Sorcery {3}{W}{W}{W}{W} (EDHREC rank 2316):
//
//	"Return all artifact and enchantment cards from your graveyard to
//	 the battlefield."
//
// The seven-mana Replenish-and-more. Every artifact and enchantment
// card in the caster's graveyard comes back under its owner's
// control — an artifact creature is an artifact and counts — all of
// them as one entry (#1867), so its own enters-tapped clause runs and
// every ETB trigger sees the whole batch (CR 603.6a). Not targeted, as printed, and
// castable with an empty graveyard to no effect.
//
// An Aura returned this way chooses what it enchants as it enters
// (CR 303.4f, #3158), one question per Aura, among the objects and
// players already there — not another permanent this spell is
// returning. One with nothing to enchant stays in the graveyard
// (CR 303.4g).
func init() {
	Register(Spec{
		OracleID:     "9584a8ae-2aba-42b8-8983-0467d6bd5698",
		Name:         "Brilliant Restoration",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b21ReturnAllArtifactAndEnchantmentCards(ctx, ctx.Controller())
		},
	})
}
