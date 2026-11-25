package cardsap

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Azra Oddsmaker — Creature — Azra Warrior {1}{B}{R}, 3/3 (Battlebond, #5529):
//
//	At the beginning of combat on your turn, you may discard a card. If you do, choose a creature. Whenever that creature deals combat damage to a player this turn, you draw two cards.
//
// Furyblade Vampire's discard offer; the chosen creature is bound as an object (id and epoch) to a repeating delayed trigger, as Captain Howler does.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "457470db-e664-4ca3-84ec-6cec2d4ae906",
		Name:         "Azra Oddsmaker",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.AtBeginningOfYourCombat("Azra Oddsmaker — you may discard a card; if you do, choose a creature",
				func(g *game.Game, item *game.StackItem) error {
					return effects.MayPay{
						Chooser:  item.Controller,
						Pay:      effects.DiscardPayment(1),
						Question: "Azra Oddsmaker — discard a card to choose a creature?",
						Then:     azraOddsmakerPaid,
					}.Apply(effects.NewContext(g, item))
				}),
		},
	})
}

var azraOddsmakerPaid = effects.PayUnlessBranches("pay-unless/azra-oddsmaker-choose",
	func(ctx *effects.Context, _ game.PayUnlessAnswer) error {
		var creatures []uuid.UUID
		for _, c := range ctx.Game.BattlefieldCardsForEffect() {
			if c.IsCreature() {
				creatures = append(creatures, c.InstanceID)
			}
		}
		if len(creatures) == 0 {
			return nil
		}
		ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
			Chooser:  ctx.Controller(),
			Source:   ctx.Source(),
			Question: "Azra Oddsmaker — choose a creature",
			Cards:    creatures,
			Min:      1,
			Max:      1,
			ThenKey:  azraOddsmakerChosen,
			LiveZone: game.ZoneBattlefield,
			Asked:    game.NewAskedItem(ctx.Item),
		})
		return nil
	}, nil)

var azraOddsmakerChosen = effects.AskedPick("choose-cards/azra-oddsmaker-chosen", func(ctx *effects.Context, r game.CardsChosen) error {
	if len(r.Picked) == 0 {
		return nil
	}
	c, ok := ctx.Game.LookupCardForEffect(r.Picked[0])
	if !ok {
		return nil
	}
	return effects.DelayedOnEvent{
		Label:      "Azra Oddsmaker — you draw two cards",
		On:         []game.EventKind{game.EventDealDamage},
		Condition:  effects.ThatCreatureHitAPlayerCondition,
		CondParams: game.EffectParams{Object: game.ObjectRef{ID: r.Picked[0], Epoch: c.ObjectEpoch}},
		Body:       azraOddsmakerDraw,
		Repeats:    true,
	}.Apply(ctx)
})

var azraOddsmakerDraw = game.SimpleDelayedBody("draw/s82-old-b205-azra-oddsmaker-two", func(g *game.Game, item *game.StackItem) error {
	return effects.DrawCards{Player: item.Controller, N: 2}.Apply(effects.NewContext(g, item))
})
