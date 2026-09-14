package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Replenish — Sorcery {3}{W} (EDHREC rank 3794):
//
//	"Return all enchantment cards from your graveyard to the
//	 battlefield. (Auras with nothing to enchant remain in your
//	 graveyard.)"
//
// The enchantress deck's mass reanimation. Every non-Aura enchantment
// card in the caster's graveyard comes back under its owner's
// control — an enchantment creature is an enchantment and counts —
// each through the ordinary reanimation path, so its own
// enters-tapped clause and every ETB trigger fire. Not targeted, as
// printed, and castable with an empty graveyard to no effect.
//
// Sandbox simplification, declared: EVERY Aura stays in the
// graveyard, not only the ones with nothing to enchant. CR 303.4f —
// the Aura's controller chooses what it enchants as it enters — has
// no prompt on the reanimation path, and Brilliant Restoration's
// posture (return it unattached, let CR 704.5n bin it) would fire
// the Aura's enters triggers for nothing. Leaving it where it is
// matches the printed parenthetical exactly for the no-legal-host
// case and is weaker than printed for the rest, never stronger.
func init() {
	Register(Spec{
		OracleID:     "523ae937-5535-490d-96ea-07f331b5e5ad",
		Name:         "Replenish",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"Auras stay in your graveyard — you don't get to choose what they enchant, so only non-Aura enchantments return."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b36ReturnAllNonAuraEnchantmentCards(ctx)
		},
	})
}
