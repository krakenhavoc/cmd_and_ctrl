package cardssy

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Tezzeret, Master of the Bridge — Legendary Planeswalker — Tezzeret {4}{U}{B},
// loyalty 5:
//
//	"Creature and planeswalker spells you cast have affinity for artifacts.
//	 +2: Tezzeret deals X damage to each opponent, where X is the number of
//	 artifacts you control. You gain X life.
//	 −3: Return target artifact card from your graveyard to your hand.
//	 −8: Exile the top ten cards of your library. Put all artifact cards from
//	 among them onto the battlefield."
//
// The affinity is Pearl-Ear's shape: a cost reduction of {1} per artifact you
// control on creature and planeswalker spells you cast. X is counted once, as
// +2 resolves, and is the damage and the life.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "d097991c-41df-47de-a439-57a0693b9581",
		Name:            "Tezzeret, Master of the Bridge",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		CostModifiers: []game.CostModifier{
			effects.CostsLessEach(effects.PermanentsYouControl(effects.Artifact()),
				"Creature and planeswalker spells you cast have affinity for artifacts.",
				effects.YourSpell(), func(q game.CostQuery) bool { return q.Card.IsCreature() || q.Card.IsPlaneswalker() }),
		},
		Activated: []effects.ActivatedAbility{
			{
				Label: "+2: Tezzeret deals X damage to each opponent, where X is the number of artifacts you control. You gain X life.",
				Cost:  effects.LoyaltyCost(2),
				Effect: func(g *game.Game, item *game.StackItem) error {
					x := 0
					for _, c := range g.BattlefieldCardsForEffect() {
						if c.Controller == item.Controller && c.IsArtifact() {
							x++
						}
					}
					if err := effects.DamageToEachOpponent(g, item, x); err != nil {
						return err
					}
					return effects.GainLife{Amount: x}.Apply(effects.NewContext(g, item))
				},
			},
			{
				Label:   "−3: Return target artifact card from your graveyard to your hand.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCardInGraveyard("target artifact card in your graveyard", effects.Artifact(), effects.YouOwn()),
				Effect:  effects.ReturnTargetedCardToHand,
			},
			{
				Label: "−8: Exile the top ten cards of your library. Put all artifact cards from among them onto the battlefield.",
				Cost:  effects.LoyaltyCost(-8),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					return effects.MillToZone{Player: item.Controller, N: 10, To: game.ZoneExile,
						ThenTail: ctx.ItemTail(tezzeretBridgeArtifacts)}.Apply(ctx)
				},
			},
		},
	})
}

// tezzeretBridgeArtifacts is −8's second half: the artifact cards among the
// exiled ones enter together. An on-disk key: never renamed, never reused.
var tezzeretBridgeArtifacts = effects.CardTail("old-b007/tezzeret-bridge-artifacts", func(ctx *effects.Context, o game.TailOutcome) error {
	var ids []uuid.UUID
	for _, id := range o.IDs {
		if c, ok := ctx.Game.LookupCardForEffect(id); ok && c.IsArtifact() {
			ids = append(ids, id)
		}
	}
	return effects.ReturnFromExileTogether{Targets: ids, Controller: ctx.Controller()}.Apply(ctx)
})
