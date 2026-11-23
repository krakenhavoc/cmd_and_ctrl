package cardskr

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Lord of the Void — Creature — Demon {4}{B}{B}{B}, 7/7 (Ravnica
// Remastered, #5948):
//
//	"Flying
//	 Whenever this creature deals combat damage to a player, exile the top
//	 seven cards of that player's library, then put a creature card from
//	 among them onto the battlefield under your control."
//
// Anzrag's Rampage's shape. The seven leave together (one CR 614 window,
// ExileTopKeyedForEffect) and the continuation is told the cards that
// reached exile; the creature card among them is picked by the Lord's
// controller (not optional while there is one), and enters under their
// control. The rest stay exiled. A library of fewer than seven exiles what
// it has; no creature among them puts nothing onto the battlefield.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "626b85f8-8889-46e6-9659-460bac79ae9b",
		Name:            "Lord of the Void",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.WheneverThisDealsCombatDamageToAPlayer("Lord of the Void — exile the top seven cards of that player's library, then put a creature card from among them onto the battlefield under your control",
				func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					victim := ctx.Trigger().Event.Target
					if g.PlayerByIDForEffect(victim) == nil {
						return nil
					}
					return g.ExileTopKeyedForEffect(victim, ctx.Controller(), 7, ctx.Tail(lordOfTheVoidExiled))
				}),
		},
	})
}

var lordOfTheVoidExiled = effects.CardTail("tail/lord-of-the-void-exiled", func(ctx *effects.Context, o game.TailOutcome) error {
	var creatures []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsCreature() {
			if z := ctx.Game.FindCardZoneForEffect(id); z != nil && z.Kind == game.ZoneExile {
				creatures = append(creatures, id)
			}
		}
	}
	if len(creatures) == 0 {
		return nil
	}
	ctx.Game.QueueChooseCardsForEffect(game.ChooseCardsPrompt{
		Chooser:  ctx.Controller(),
		Source:   ctx.Source(),
		Question: "Lord of the Void — put a creature card exiled this way onto the battlefield under your control",
		Cards:    creatures,
		Min:      1,
		Max:      1,
		LiveZone: game.ZoneExile,
		ThenKey:  lordOfTheVoidPicked,
	})
	return nil
})

var lordOfTheVoidPicked = game.RegisterChooseCardsThen("choose-cards/lord-of-the-void-creature", func(g *game.Game, r game.CardsChosen) error {
	if len(r.Picked) != 1 {
		return nil
	}
	id := r.Picked[0]
	if z := g.FindCardZoneForEffect(id); z == nil || z.Kind != game.ZoneExile {
		return nil
	}
	return g.PutOntoBattlefieldTogetherKeyedForEffect([]game.BatchEntry{{CardID: id, From: game.ZoneExile}},
		game.ZoneEntryOptions{Controller: r.Chooser}, game.KeyedTail{})
})
