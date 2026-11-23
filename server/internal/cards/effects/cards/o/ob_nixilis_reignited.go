package cardso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Ob Nixilis Reignited — Legendary Planeswalker — Nixilis {3}{B}{B},
// starting loyalty 5:
//
//	"+1: You draw a card and you lose 1 life.
//	 −3: Destroy target creature.
//	 −8: Target opponent gets an emblem with "Whenever a player draws a
//	 card, you lose 2 life.""
//
// The −8 targets an opponent and that player gets the emblem (CR 114.2:
// the emblem's owner is the player who got it), so "you" in the emblem is
// the targeted opponent. It fires once per card drawn by any player,
// including its owner and the draws it is itself part of.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "65481ae1-54b2-44d2-91e1-20cd486386da",
		Name:            "Ob Nixilis Reignited",
		Completeness:    effects.CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &effects.EmblemSpec{
			Label: "Ob Nixilis Reignited emblem",
			Text:  "Whenever a player draws a card, you lose 2 life.",
			Triggered: []game.TriggeredAbility{{
				Watches: []game.EventKind{game.EventDrawCard},
				AppliesTo: func(ev game.Event, _ *game.Card, _ game.Characteristic, _ *game.Game) bool {
					return ev.Actor != uuid.Nil
				},
				Key: "Ob Nixilis Reignited emblem — you lose 2 life",
				Effect: func(g *game.Game, item *game.StackItem) error {
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -2)
				},
			}},
		},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "+1: You draw a card and you lose 1 life.",
				Cost:    effects.LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect:  effects.B36DrawAndLoseOne,
			},
			{
				Label:   "−3: Destroy target creature.",
				Cost:    effects.LoyaltyCost(-3),
				Targets: effects.TargetCreature("target creature"),
				Effect:  effects.DestroyFirstLegalCardTarget,
			},
			{
				Label:   "−8: Target opponent gets an emblem with \"Whenever a player draws a card, you lose 2 life.\"",
				Cost:    effects.LoyaltyCost(-8),
				Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
				Effect:  effects.TargetOpponentGetsAnEmblem,
			},
		},
	})
}
