package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Brilliant Restoration — Sorcery {3}{W}{W}{W}{W} (EDHREC rank 2316):
//
//	"Return all artifact and enchantment cards from your graveyard to
//	 the battlefield."
//
// The seven-mana Replenish-and-more. Every artifact and enchantment
// card in the caster's graveyard comes back under its owner's
// control — an artifact creature is an artifact and counts — each
// through the ordinary reanimation path, so its own enters-tapped
// clause and every ETB trigger fire. Not targeted, as printed, and
// castable with an empty graveyard to no effect.
//
// Sandbox simplification, declared: an Aura returned this way comes
// back UNATTACHED, and CR 704.5m then puts it straight into the
// graveyard. CR 303.4f is the clause that is missing — it lets the
// Aura's controller choose what it enchants as it enters, and the
// reanimation path has no such prompt. So the Aura goes back to the
// graveyard it came from instead of enchanting something. Weaker than
// printed (the Aura does nothing), never stronger.
//
// It used to SIT on the battlefield forever instead, which was worse
// than weak: an Aura attached to nothing is a board state CR 704.5m
// forbids and no sequence of legal plays can reach. That branch of
// the state-based action landed in the S24 tail.
func init() {
	Register(Spec{
		OracleID:     "9584a8ae-2aba-42b8-8983-0467d6bd5698",
		Name:         "Brilliant Restoration",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"An Aura returned this way comes back unattached and is put into the graveyard — you don't get to choose what it enchants."},
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return b21ReturnAllArtifactAndEnchantmentCards(ctx, ctx.Controller())
		},
	})
}
