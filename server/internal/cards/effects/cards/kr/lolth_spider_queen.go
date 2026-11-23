package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Lolth, Spider Queen — Legendary Planeswalker — Lolth {3}{B}{B}, loyalty 4
// (Adventures in the Forgotten Realms, S82 batch 15, #5499):
//
//	"Whenever a creature you control dies, put a loyalty counter on Lolth.
//	 0: You draw a card and you lose 1 life.
//	 −3: Create two 2/1 black Spider creature tokens with reach and menace.
//	 −8: You get an emblem with 'Whenever an opponent is dealt combat damage
//	     by one or more creatures you control, if that player lost less than
//	     8 life this turn, they lose life equal to the difference.'"
//
// The emblem's trigger is one per opponent per combat damage step (CR 603.2c,
// OncePerBatchPerPlayer). Its intervening "if" (CR 603.4) reads the life the
// player has lost this turn, combat damage included (CR 119.3, the turn
// tally), when it would trigger and again as it resolves; the loss is 8
// minus that, read at resolution. The emblem is the source.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7a8edf45-206e-4d0e-bd69-06f1ca70d17c",
		Name:            "Lolth, Spider Queen",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 4,
		Emblem: &effects.EmblemSpec{
			Label: "Lolth, Spider Queen emblem",
			Text:  "Whenever an opponent is dealt combat damage by one or more creatures you control, if that player lost less than 8 life this turn, they lose life equal to the difference.",
			Triggered: []game.TriggeredAbility{
				effects.OncePerBatchPerPlayer(effects.On(game.EventDealDamage,
					func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
						return ev.Target != source.Controller && effects.CombatDamageToPlayerBy(ev, source.Controller, g) &&
							g.TurnTallyFor(ev.Target).LifeLost < 8
					},
					"Lolth, Spider Queen emblem — that player loses life equal to 8 minus the life they lost this turn",
					lolthEmblemDrain)),
			},
		},
		Triggered: []game.TriggeredAbility{
			effects.WheneverACreatureYouControlDies("Lolth — put a loyalty counter on Lolth",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.AddCounter{Target: item.SourceCardID, Kind: game.CounterLoyalty, N: 1}.Apply(ctx)
				}),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "0: You draw a card and you lose 1 life.",
				Cost:    effects.LoyaltyCost(0),
				Purpose: game.Purpose{Draws: 1},
				Effect:  effects.Do(effects.DrawCards{N: 1}, effects.GainLife{Amount: -1}),
			},
			{
				Label:  "−3: Create two 2/1 black Spider creature tokens with reach and menace.",
				Cost:   effects.LoyaltyCost(-3),
				Effect: effects.Do(effects.CreateToken{Template: effects.TokenCard(effects.AfrB15SpiderToken), N: 2}),
			},
			{
				Label:  "−8: You get an emblem with \"Whenever an opponent is dealt combat damage by one or more creatures you control, if that player lost less than 8 life this turn, they lose life equal to the difference.\"",
				Cost:   effects.LoyaltyCost(-8),
				Effect: effects.Do(effects.CreateEmblem{}),
			},
		},
	})
}

// lolthEmblemDrain makes the damaged player lose 8 minus the life they have
// lost this turn, rechecking the intervening "if" as the ability resolves.
func lolthEmblemDrain(g *game.Game, item *game.StackItem) error {
	ctx := effects.NewContext(g, item)
	player := ctx.Trigger().Event.Target
	diff := 8 - g.TurnTallyFor(player).LifeLost
	if diff <= 0 {
		return nil
	}
	return effects.GainLife{Player: player, Amount: -diff}.Apply(ctx)
}
