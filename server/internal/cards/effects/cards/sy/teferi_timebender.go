package cardssy

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Teferi, Timebender — Legendary Planeswalker — Teferi {4}{W}{U}, loyalty 5
// (Dominaria, S82 OLD batch 215, #5530):
//
//	"+2: Untap up to one target artifact or creature.
//	 −3: You gain 2 life and draw two cards.
//	 −9: Take an extra turn after this one."
//
// The +2 has up to one target, so it can be activated with none (CR 115.4).
// The −3 gains the life, then draws. The −9 queues one extra turn after the
// current one (CR 500.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "761bb894-d3f9-4248-af76-1559c502e342",
		Name:            "Teferi, Timebender",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+2: Untap up to one target artifact or creature.",
				Cost:    effects.LoyaltyCost(2),
				Targets: effects.TargetPermanent("up to one target artifact or creature", effects.Or(effects.Artifact(), effects.Creature())).WithCount(0, 1),
				Effect:  effects.UntapEachLegalTarget,
			},
			{
				Label:   "−3: You gain 2 life and draw two cards.",
				Cost:    effects.LoyaltyCost(-3),
				Purpose: game.Purpose{Draws: 2, LifeGain: 2},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := effects.NewContext(g, item)
					if err := (effects.GainLife{Player: item.Controller, Amount: 2}).Apply(ctx); err != nil {
						return err
					}
					return effects.DrawCards{Player: item.Controller, N: 2}.Apply(ctx)
				},
			},
			{
				Label:   "−9: Take an extra turn after this one.",
				Cost:    effects.LoyaltyCost(-9),
				Purpose: game.Purpose{ExtraTurns: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					return effects.TakeExtraTurn{Label: "Teferi, Timebender — extra turn"}.Apply(effects.NewContext(g, item))
				},
			},
		},
	})
}
