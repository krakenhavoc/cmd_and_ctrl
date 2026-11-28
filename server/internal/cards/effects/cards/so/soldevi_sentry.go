package cardsso

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Soldevi Sentry — Artifact Creature — Soldier {1}, 1/1 (Alliances, #5634):
//
//	"{1}: Choose target opponent. Regenerate this creature. When it
//	 regenerates this way, that player may draw a card."
//
// The shield (CR 701.19a) remembers the opponent the activation named
// (Game.RegenerateWithOpponentMayDrawForEffect, Card.RegenerationDrawers): when
// it is used up, the regeneration is announced with that player, and the
// printed triggered ability lets THAT player (not the Sentry's controller) draw
// a card. Each activation's opponent rides its own shield, so two shields made
// in response to one another pay out to the players that named them.
//
// No simplification.
func init() {
	const name = "Soldevi Sentry"
	t := effects.On(game.EventRegenerated, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
		return ev.Kind == game.EventRegenerated && ev.CardID == source.InstanceID && ev.Target != uuid.Nil
	}, name+" — that player may draw a card", func(g *game.Game, item *game.StackItem) error {
		if item.Params.Player == uuid.Nil {
			return nil
		}
		return effects.MayChoice{
			Player:   item.Params.Player,
			Question: name + " — draw a card?",
			Draws:    1,
			YesLabel: "Draw",
			NoLabel:  "Decline",
			Then:     effects.OldB814MayDraw,
		}.Apply(effects.NewContext(g, item))
	})
	t.Build = func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) *game.StackItem {
		item := game.NewTriggeredItem(source, name+" — that player may draw a card")
		item.Params.Player = ev.Target
		return item
	}
	effects.Register(effects.Spec{
		OracleID:     "d2a61132-d881-4c80-92f1-a151d835832a",
		Name:         name,
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}: Choose target opponent. Regenerate this creature. When it regenerates this way, that player may draw a card.",
			Purpose: game.Purpose{Answers: game.AnswerSurvive},
			Cost:    effects.ManaCost("{1}"),
			Targets: effects.TargetPlayer("target opponent", effects.Opponent()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				if !effects.SourceIsStillThisPermanent(g, item) {
					return nil
				}
				ctx := effects.NewContext(g, item)
				opp := effects.TargetedPlayer(ctx)
				if opp == uuid.Nil {
					return nil
				}
				return effects.NothingIfGone(g.RegenerateWithOpponentMayDrawForEffect(item.SourceCardID, opp))
			},
		}},
		Triggered: []game.TriggeredAbility{t},
	})
}
