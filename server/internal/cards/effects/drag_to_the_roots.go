package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Drag to the Roots — Instant {2}{B}{G} (EDHREC rank 4192):
//
//	"Delirium — This spell costs {2} less to cast as long as there are
//	 four or more card types among cards in your graveyard.
//	 Destroy target nonland permanent."
//
// Unconditional instant-speed removal that a graveyard deck casts for
// two. In Golgari the delirium count is usually live by turn four, so
// the card is really a two-mana Beast Within without the drawback.
//
// The removal half is the whole card that works, and it is as plain
// as removal gets: a single nonland permanent target, destroyed.
// Indestructible and regeneration are the engine's business.
//
// DECLARED SIMPLIFICATION — NO DELIRIUM DISCOUNT. The card always
// costs {2}{B}{G}. Delirium is a SelfCostModifiers clause and the
// slot exists, but the condition it needs — "four or more card types
// among cards in your graveyard" — is the delirium count, which is a
// read the catalog has no shared body for and which several other
// cards in this roadmap want built once rather than five times
// (#746). Shipping a hand-rolled count here would be the fifth
// slightly-different one.
//
// The effect is that the spell costs two more than printed, which is
// strictly worse and therefore shippable (#259). Nothing about the
// removal changes.
func init() {
	Register(Spec{
		OracleID:     "ef4c478c-7019-4ec0-8edf-2a3078a8e97a",
		Name:         "Drag to the Roots",
		Completeness: CompletenessCaveats,
		Caveats:      []string{"The delirium discount is not implemented — the spell always costs {2}{B}{G}, never {B}{G}, however many card types are in your graveyard."},
		Targets:      TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
