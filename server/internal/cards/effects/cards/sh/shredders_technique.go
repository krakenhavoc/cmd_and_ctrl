package cardssh

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Shredder's Technique — Sorcery {2}{B} (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"Sneak {B} (You may cast this spell for {B} if you also return an
//	 unblocked attacker you control to hand during the declare blockers
//	 step.)
//	 Destroy target creature or enchantment. If an enchantment was
//	 destroyed this way, you lose 2 life."
//
// Sneak is the shared alternative cost (sneak.go, CR 702.190). The target
// is Withering Torment's. The life loss reads what was actually destroyed
// (an indestructible or regenerated target was not), as it last existed on
// the battlefield, so an enchantment creature costs 2 life and a plain
// creature costs nothing. The destroy is the keyed door
// (DestroyPermanentsKeyedForEffect), so a replacement's question in the
// middle of it is a restore point.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "d92714c0-4e22-47f9-a4d4-405894e734d0",
		Name:             "Shredder's Technique",
		Completeness:     effects.CompletenessFull,
		Purpose:          game.Purpose{Answers: game.AnswerRemove},
		AlternativeCosts: []game.AlternativeCost{effects.Sneak("{B}")},
		Targets:          effects.TargetPermanent("target creature or enchantment", effects.Or(effects.Creature(), effects.Enchantment())),
		OnResolve: func(_ *game.StackItem, ctx *effects.Context) error {
			id, ok := effects.B16FirstLegalTargetCard(ctx)
			if !ok {
				return nil
			}
			return ctx.Game.DestroyPermanentsKeyedForEffect([]uuid.UUID{id}, ctx.Tail(shreddersTechniqueLoss))
		},
	})
}

// shreddersTechniqueLoss is "if an enchantment was destroyed this way, you
// lose 2 life", keyed. An on-disk key: never renamed, never reused.
var shreddersTechniqueLoss = effects.MassTail("mass/shredders-technique-enchantment-loss", func(ctx *effects.Context, destroyed []game.Card, _ int) error {
	for _, c := range destroyed {
		if c.IsEnchantment() {
			return ctx.Game.ChangePlayerLifeForEffect(ctx.Source(), ctx.Controller(), -2)
		}
	}
	return nil
})
