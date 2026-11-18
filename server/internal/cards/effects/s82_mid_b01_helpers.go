package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// s82MidB01Return bounces the creature Loyal Gryff's controller picked.
// It is gone from the battlefield (or was never picked) means nothing happens.
var s82MidB01Return = RegisterPermanentsPick("prompt-run/loyal-gryff-return", PermanentsDef{
	Then: func(ctx *Context, picked game.PromptedPicks) error {
		for _, id := range picked.Cards() {
			if !onBattlefield(ctx.Game, id) {
				continue
			}
			return BounceToHand{Target: id}.Apply(ctx)
		}
		return nil
	},
})

// chooseAnotherCreatureYouControl is the resolution-time pick of "you may
// return another creature you control to its owner's hand" (First Responder,
// Loyal Gryff): nothing targets, zero is the "you may" decline, and the
// source itself is never offered.
func chooseAnotherCreatureYouControl(g *game.Game, item *game.StackItem, question string, keyed PermanentsThen) error {
	self := item.SourceCardID
	return ChoosePermanents{
		Question: question,
		Candidates: func(g *game.Game, of uuid.UUID) ([]uuid.UUID, int, int) {
			var out []uuid.UUID
			for _, c := range g.BattlefieldCardsForEffect() {
				if c.Controller == of && c.IsCreature() && c.InstanceID != self {
					out = append(out, c.InstanceID)
				}
			}
			return out, 0, 1
		},
		Keyed: keyed,
	}.Apply(NewContext(g, item))
}
