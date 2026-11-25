package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Golos, Tireless Pilgrim — Legendary Artifact Creature — Scout {5}, 3/5 (Core
// Set 2020, #5518):
//
//	"When Golos enters, you may search your library for a land card, put that
//	 card onto the battlefield tapped, then shuffle.
//	 {2}{W}{U}{B}{R}{G}: Exile the top three cards of your library. You may
//	 play them this turn without paying their mana costs."
//
// The search is optional (the player may decline the card and the shuffle).
// The activated ability is an impulse exile of three cards with a free play
// permission that lasts the turn: "play", so a land among them may be played
// as the turn's land drop, and a spell is cast without paying its mana cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "efe9c3e0-7315-408e-9903-9d25a681039b",
		Name:         "Golos, Tireless Pilgrim",
		Completeness: effects.CompletenessFull,
		Purpose:      game.Purpose{Lands: 1},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Golos, Tireless Pilgrim — you may search for a land card, put it onto the battlefield tapped",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:        item.Controller,
						Predicate:     effects.IsLandCard,
						Dest:          game.ZoneBattlefield,
						Limit:         1,
						Shuffle:       true,
						TappedOnEntry: true,
						Optional:      true,
						Reason:        "Golos, Tireless Pilgrim — a land card onto the battlefield tapped",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W}{U}{B}{R}{G}: Exile the top three cards of your library. You may play them this turn without paying their mana costs.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{W}{U}{B}{R}{G}"),
			Effect: func(g *game.Game, item *game.StackItem) error {
				return effects.ExileTopWithPermission{From: item.Controller, GrantTo: item.Controller, N: 3, Free: true}.Apply(effects.NewContext(g, item))
			},
		}},
	})
}
