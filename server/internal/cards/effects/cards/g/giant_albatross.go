package cardsg

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Giant Albatross — Creature — Bird {1}{U}, 1/1 (Homelands, S82 OLD b827,
// #5635):
//
//	"Flying
//	 When this creature dies, you may pay {1}{U}. If you do, for each creature
//	 that dealt damage to this creature this turn, destroy that creature unless
//	 its controller pays 2 life. A creature destroyed this way can't be
//	 regenerated."
//
// Flying is a printed keyword. The dies trigger is a may-pay: nothing is read
// until the {1}{U} has been paid, then every creature that dealt damage to the
// Albatross this turn (the turn's damage record, read as the object that has
// just died, CR 400.7) that is still the same creature on the battlefield is
// put to its controller as a pay-2-life prompt of its own; a controller who
// declines, or has less than 2 life, loses the creature, with regeneration
// refused (CR 701.19c). A creature that is already gone is not asked about.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "8b59cacb-46ff-4400-a7f6-b258c81a9b14",
		Name:            "Giant Albatross",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisDies("Giant Albatross — you may pay {1}{U}; destroy each creature that dealt damage to it this turn unless its controller pays 2 life",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Cost:     "{1}{U}",
						Question: "Giant Albatross — pay {1}{U}? (if you do, each creature that dealt damage to it this turn is destroyed unless its controller pays 2 life)",
						Then:     giantAlbatrossPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var (
	// giantAlbatrossPaid is the paid branch: one pay-2-life-or-be-destroyed
	// prompt per creature that dealt damage to the Albatross. An on-disk key:
	// never renamed, never reused.
	giantAlbatrossPaid = effects.PayUnlessBranches("pay-unless/old-b827-giant-albatross-paid",
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			pay := effects.LifePayment(2)
			for _, ref := range ctx.Game.DamageDealersToPermanentThisTurnForEffect(r.Source) {
				c, ok := ctx.Game.LookupCardForEffect(ref.ID)
				if !ok || c.ObjectEpoch != ref.Epoch || !effects.OnBattlefield(ctx.Game, ref.ID) || !c.IsCreature() {
					continue
				}
				if err := ctx.Game.QueuePayUnlessPromptKeyedForEffect(game.PayUnlessKeyedPrompt{
					Chooser:  c.Controller,
					Source:   ctx.Source(),
					Cost:     pay.Words,
					Question: "Giant Albatross — pay 2 life or " + c.Name + " is destroyed?",
					Action:   pay.Action,
					Then:     giantAlbatrossDeclined,
					Carry:    []uuid.UUID{ctx.Controller()},
					Objects:  []game.ObjectRef{ref},
					Asked:    game.NewAskedItem(ctx.Item),
				}); err != nil {
					return err
				}
			}
			return nil
		}, nil)
	// giantAlbatrossDeclined is the creature's controller not paying: the
	// creature, if it is still the object that dealt the damage, is destroyed
	// and can't be regenerated. An on-disk key: never renamed, never reused.
	giantAlbatrossDeclined = effects.PayUnlessBranches("pay-unless/old-b827-giant-albatross-destroy", nil,
		func(ctx *effects.Context, r game.PayUnlessAnswer) error {
			if len(r.Objects) == 0 {
				return nil
			}
			ref := r.Objects[0]
			c, ok := ctx.Game.LookupCardForEffect(ref.ID)
			if !ok || c.ObjectEpoch != ref.Epoch || !effects.OnBattlefield(ctx.Game, ref.ID) {
				return nil
			}
			return effects.DestroyTarget{Target: ref.ID, CantBeRegenerated: true}.Apply(ctx)
		})
)
