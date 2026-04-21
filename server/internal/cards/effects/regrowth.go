package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Regrowth — "Return target card from your graveyard to your hand."
//
// S14 sandbox simplification: no graveyard-picker UI yet. The
// primitive auto-picks the MOST RECENTLY added card in the
// controller's graveyard (top of the pile). Players who want a
// specific card can rearrange their graveyard through the existing
// drag-and-drop affordance before casting. A proper "target a card
// in graveyard" picker is S20 smart-cast territory. TargetMode
// stays empty so the client skips the cast-time prompt.
//
// Regrowth itself does not sit in the caster's graveyard at
// resolution time — OnResolve runs before zone routing, so the
// spell is still on the stack when we read the graveyard pile.
// This matters: otherwise Regrowth would tend to return itself.
func init() {
	Register(Spec{
		OracleID: "e6e4a8bd-5c40-4654-8de1-0da9afed90fd",
		Name:     "Regrowth",
		OnResolve: func(item *game.StackItem, ctx *Context) error {
			controller := ctx.PlayerByID(ctx.Controller())
			if controller == nil || controller.Graveyard.Size() == 0 {
				return nil
			}
			top := controller.Graveyard.Cards[controller.Graveyard.Size()-1].InstanceID
			return ReturnFromGraveyard{Target: top, Dest: game.ZoneHand}.Apply(ctx)
		},
	})
}
