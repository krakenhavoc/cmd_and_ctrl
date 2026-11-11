package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Retether — Sorcery {3}{W} (Wilds of Eldraine Commander, #5368):
//
//	"Return each Aura card from your graveyard to the battlefield. Only
//	 creatures can be enchanted this way. (Aura cards that can't enchant a
//	 creature on the battlefield remain in your graveyard.)"
//
// Replenish's shape: every Aura card is read once before anything moves and
// all of them enter as one simultaneous entry, so every enters trigger sees
// the whole batch and no returning Aura can enchant another (CR 303.4f).
// Each Aura chooses its host as it enters among the CREATURES its own
// enchant clause admits (ZoneEntryOptions.AuraHostsCreaturesOnly): hexproof
// does not stop the choice but protection does, a creature card in a
// graveyard is not a host (Animate Dead stays), and an Aura with no host
// stays in the graveyard (CR 303.4g). The Auras return under their owner's
// control, which is mine: they are my graveyard's.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b693c24f-07d9-40fa-aa4a-8f7ba61f3452",
		Name:         "Retether",
		Completeness: CompletenessFull,
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			return ReturnFromGraveyardTogether{
				Targets: graveyardCardIDs(ctx, ctx.Controller(), func(c game.Card) bool {
					return c.IsAura()
				}),
				AuraHostsCreaturesOnly: true,
			}.Apply(ctx)
		},
	})
}
