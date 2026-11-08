package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Break the Spell — Instant {W} (Wilds of Eldraine, #4471):
//
//	"Destroy target enchantment. If a permanent you controlled or a token
//	 was destroyed this way, draw a card."
//
// Shredder's Technique's shape: the destroy is the keyed door
// (DestroyPermanentsKeyedForEffect), and the draw reads what was actually
// destroyed, as it last existed on the battlefield (CR 603.10). An
// indestructible or regenerated enchantment was not destroyed, so it draws
// nothing; your own enchantment, or anyone's token enchantment (a Role),
// draws a card.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "b9d51e07-851f-4b7b-bffe-7c1efff0714c",
		Name:         "Break the Spell",
		Completeness: CompletenessFull,
		Purpose:      game.Purpose{Answers: game.AnswerRemove},
		Targets:      TargetPermanent("target enchantment", Enchantment()),
		OnResolve: func(_ *game.StackItem, ctx *Context) error {
			id, ok := b16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, ctx.Tail(breakTheSpellDraw))
		},
	})
}

// breakTheSpellDraw is "if a permanent you controlled or a token was
// destroyed this way, draw a card", keyed. An on-disk key: never renamed,
// never reused.
var breakTheSpellDraw = MassTail("mass/break-the-spell-draw", func(ctx *Context, destroyed []game.Card, _ int) error {
	for _, c := range destroyed {
		if c.Controller == ctx.Controller() || c.IsToken() {
			return DrawCards{Player: ctx.Controller(), N: 1}.Apply(ctx)
		}
	}
	return nil
})
