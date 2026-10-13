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
// The discount is a SelfCostModifiers clause reading the shared delirium
// count (b16CardTypesInGraveyard, via YouHaveDelirium), so it is priced
// the same way for the cast button, the auto-tapper and the bot.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ef4c478c-7019-4ec0-8edf-2a3078a8e97a",
		Name:         "Drag to the Roots",
		Completeness: CompletenessFull,
		SelfCostModifiers: []game.CostModifier{
			CostsLess(2, "Delirium — This spell costs {2} less to cast as long as there are four or more card types among cards in your graveyard.", YouHaveDelirium()),
		},
		Targets: TargetPermanent("target nonland permanent", Nonland()),
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			if len(item.Targets) == 0 || item.Targets[0].Kind != game.TargetCard {
				return nil
			}
			return DestroyTarget{Target: item.Targets[0].ID}.Apply(ctx)
		},
	})
}
