package cardso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Olivia, Mobilized for War — Legendary Creature — Vampire Knight {1}{B}{R},
// 3/3 flying (Secret Lair Drop):
//
//	"Whenever another creature you control enters, you may discard a card. If
//	 you do, put a +1/+1 counter on that creature, it gains haste until end of
//	 turn, and it becomes a Vampire in addition to its other types."
//
// The entering creature rides the prompt as a carried object with its epoch,
// so a creature that has left (or left and returned, CR 400.7) is not touched.
// The Vampire type has no duration: it is a permanent type-adding effect.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "1f8d387e-e4bb-4e76-a173-f46f7b107b95",
		Name:            "Olivia, Mobilized for War",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AnotherCreatureEnteredUnderYourControl,
				"Olivia, Mobilized for War — you may discard a card; if you do, +1/+1 counter, haste and Vampire on that creature",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					id := ctx.Trigger().Event.CardID
					c, ok := g.LookupCardForEffect(id)
					if !ok {
						return nil
					}
					return effects.MayPay{
						Chooser:      ctx.Controller(),
						Question:     "Olivia, Mobilized for War — discard a card to put a +1/+1 counter on the entering creature, give it haste and make it a Vampire?",
						Pay:          effects.DiscardPayment(1),
						Then:         oliviaMobilizedPaid,
						CarryObjects: []game.ObjectRef{{ID: id, Epoch: c.ObjectEpoch}},
					}.Apply(ctx)
				}),
		},
	})
}

var oliviaMobilizedPaid = effects.PayUnlessBranches("pay-unless/olivia-mobilized-for-war-discarded",
	func(ctx *effects.Context, r game.PayUnlessAnswer) error {
		if len(r.Objects) != 1 {
			return nil
		}
		return effects.OldB008OliviaBoost(ctx, r.Objects[0].ID, r.Objects[0].Epoch)
	}, nil)
