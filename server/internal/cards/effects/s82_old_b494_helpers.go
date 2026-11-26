package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82_old_b494_helpers.go — shared pieces of S82 OLD batch 494 (Zendikar
// enters triggers, #5583).

// Old494AlliesYouControl lists the Allies `player` controls, read through the
// layered view and the source's text (CR 612.2).
func Old494AlliesYouControl(ctx *Context, player uuid.UUID) []uuid.UUID {
	ally := ctx.TextSubtype("Ally")
	var ids []uuid.UUID
	for _, c := range ctx.Game.BattlefieldCardsForEffect() {
		if c.Controller == player && c.HasSubtype(ally) {
			ids = append(ids, c.InstanceID)
		}
	}
	return ids
}

// Old494ReturnTwoLands is Khalni Gem's "return two lands you control to their
// owner's hand": the controller chooses two (all of them with fewer than two)
// on resolution, a choice and not a target.
func Old494ReturnTwoLands(g *game.Game, item *game.StackItem) error {
	return ChoosePermanents{
		Question: "Khalni Gem — return two lands you control to their owner's hand",
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsLand() {
					out = append(out, c.InstanceID)
				}
			}
			n := min(2, len(out))
			return out, n, n
		},
		Keyed: BouncedChosen,
	}.Apply(NewContext(g, item))
}

// Old494AlliesGainProtection is Kabira Evangel's "you may choose a color. If
// you do, Allies you control gain protection from the chosen color until end
// of turn": the Allies are read as the ability resolves, then the colour is
// asked. The "may" is the trigger's.
func Old494AlliesGainProtection(g *game.Game, item *game.StackItem) error {
	ctx := NewContext(g, item)
	ids := Old494AlliesYouControl(ctx, item.Controller)
	if len(ids) == 0 {
		return nil
	}
	then := ctx.ItemTail(RadianceProtectionFromChosenColor)
	then.Carry = ids
	then.Words = []string{"Kabira Evangel"}
	ChooseColorTail(game.ColorForProtection, g, item.Controller, ctx.Source(), "Kabira Evangel — choose a color", then)
	return nil
}
