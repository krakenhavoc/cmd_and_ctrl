package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Helpers shared by cards from more than one batch in S82 roll-up 41 (#5948).

// R41ExileEnchanted is an Aura's "exile enchanted creature": whatever the
// Aura (the ability's source) is attached to as the ability resolves. Caught
// in the Brights and Weight of Conscience use it.
func R41ExileEnchanted(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	aura, ok := ctx.SourcePermanent()
	if !ok || aura.AttachedTo.Kind != game.TargetCard {
		return nil
	}
	return ExileTarget{Target: aura.AttachedTo.ID}.Apply(ctx)
}
