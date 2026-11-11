package effects

import (
	"github.com/google/uuid"

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
	Register(Spec{
		OracleID:        "65481ae1-54b2-44d2-91e1-20cd486386da",
		Name:            "Ob Nixilis Reignited",
		Completeness:    CompletenessFull,
		StartingLoyalty: 5,
		Emblem: &EmblemSpec{
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
		Activated: []ActivatedAbility{
			{
				Label:   "+1: You draw a card and you lose 1 life.",
				Cost:    LoyaltyCost(1),
				Purpose: game.Purpose{Draws: 1},
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					if err := (DrawCards{N: 1}).Apply(ctx); err != nil {
						return err
					}
					return g.ChangePlayerLifeForEffect(item.SourceCardID, item.Controller, -1)
				},
			},
			{
				Label:   "−3: Destroy target creature.",
				Cost:    LoyaltyCost(-3),
				Targets: TargetCreature("target creature"),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						return DestroyTarget{Target: t.ID}.Apply(ctx)
					}
					return nil
				},
			},
			{
				Label:   "−8: Target opponent gets an emblem with \"Whenever a player draws a card, you lose 2 life.\"",
				Cost:    LoyaltyCost(-8),
				Targets: TargetPlayer("target opponent", Opponent()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					ctx := NewContext(g, item)
					for _, t := range ctx.LegalTargets() {
						if t.Kind == game.TargetPlayer {
							return CreateEmblem{Player: t.ID}.Apply(ctx)
						}
					}
					return nil
				},
			},
		},
	})
}
