package cardsa

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Aerial Surveyor — Artifact — Vehicle {2}{W}, 3/4 (Kamigawa: Neon Dynasty
// Commander, S82):
//
//	"Flying
//	 Whenever this Vehicle attacks, if defending player controls more lands
//	 than you, search your library for a basic Plains card, put it onto the
//	 battlefield tapped, then shuffle.
//	 Crew 2"
//
// Keeper of the Accord's land comparison with the defending player read off
// the attack (a planeswalker or battle attacked means its controller). The
// intervening "if" (CR 603.4) is checked as the Vehicle attacks and again as
// the trigger resolves. The search is not optional, so it is not asked.
//
// No simplification.
func init() {
	const label = "Aerial Surveyor — search for a basic Plains, put it onto the battlefield tapped, then shuffle"
	moreLands := func(g *game.Game, you, defender uuid.UUID) bool {
		return defender != uuid.Nil && defender != you &&
			effects.B11OpponentControlsMoreThanYou(g, you, defender, game.Card.IsLand)
	}
	effects.Register(effects.Spec{
		OracleID:        "92c9b2fd-95d6-4937-aee2-67b4bb8ca29d",
		Name:            "Aerial Surveyor",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"flying"},
		Triggered: []game.TriggeredAbility{
			effects.TriggerWithPurpose(effects.On(game.EventAttack, func(ev game.Event, source *game.Card, _ game.Characteristic, g *game.Game) bool {
				return effects.AttackDeclared(ev, source) && moreLands(g, source.Controller, effects.B17DefendingPlayer(g, ev))
			}, label, func(g *game.Game, item *game.StackItem) error {
				if item.Trigger == nil || !moreLands(g, item.Controller, effects.B17DefendingPlayer(g, item.Trigger.Event)) {
					return nil
				}
				return effects.SearchLibrary{
					Player: item.Controller,
					Predicate: func(c game.Card) bool {
						return effects.IsBasicLand(c) && c.HasSubtypeWord("Plains")
					},
					Dest:          game.ZoneBattlefield,
					Limit:         1,
					Reveal:        true,
					Shuffle:       true,
					TappedOnEntry: true,
					Reason:        "Aerial Surveyor — a basic Plains card, onto the battlefield tapped",
				}.Apply(effects.NewContext(g, item))
			}), game.Purpose{Lands: 1}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "Crew 2",
			Cost:    effects.CrewCost(2),
			Purpose: game.Purpose{Answers: game.AnswerAnimate},
			Effect:  effects.CrewEffect("Aerial Surveyor"),
		}},
	})
}
