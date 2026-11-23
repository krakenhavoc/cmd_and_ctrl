package cardsr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Raze to the Ground — Sorcery for {2}{R} (The Brothers' War):
//
//	"This spell can't be countered.
//	Destroy target artifact. If its mana value was 1 or less, draw a card."
//
// The mana value is the target's as the spell resolves, read before it is
// destroyed (CR 608.2h). The card does not depend on the destruction: an
// indestructible artifact with a small mana value still draws, since the text
// says "if its mana value was", not "if you do". A target that has become
// illegal makes the spell do nothing (CR 608.2b).
func init() {
	effects.Register(effects.Spec{
		OracleID:        "efe24eed-e1e2-4807-ae61-2e5be529ae3c",
		Name:            "Raze to the Ground",
		Purpose:         game.Purpose{Answers: game.AnswerRemove, Draws: 1},
		Completeness:    effects.CompletenessFull,
		CantBeCountered: true,
		Targets:         effects.TargetPermanent("target artifact", effects.Artifact()),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			c, ok := ctx.Game.LookupCardForEffect(id)
			if !ok {
				return nil
			}
			small := c.ManaValue() <= 1
			if err := (effects.DestroyTarget{Target: id}).Apply(ctx); err != nil {
				return err
			}
			if !small {
				return nil
			}
			return effects.DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		},
	})
}
