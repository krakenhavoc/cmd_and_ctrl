package cardsh

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Hedron Alignment — Enchantment {2}{U} (Oath of the Gatewatch, #5545):
//
//	"Hexproof
//	 At the beginning of your upkeep, you may reveal your hand. If you do, you
//	 win the game if you own a card named Hedron Alignment in exile, in your
//	 hand, in your graveyard, and on the battlefield.
//	 {1}{U}: Scry 1."
//
// The upkeep trigger is optional; the hand is revealed as it resolves, then the
// four zones are read: exile (any exiled card you own), your hand, your
// graveyard, and the battlefield (any permanent you own, whoever controls it).
// Each needs its own card, so the enchantment on the battlefield counts for the
// battlefield and four copies are needed in all.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "7cdc9df9-ac6c-49c7-a521-921588696b17",
		Name:            "Hedron Alignment",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"hexproof"},
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.On(game.EventBeginUpkeep, func(ev game.Event, source *game.Card, _ game.Characteristic, _ *game.Game) bool {
				return ev.Actor == source.Controller
			}, "Hedron Alignment — you may reveal your hand; if you own a card named Hedron Alignment in exile, in your hand, in your graveyard and on the battlefield, you win the game",
				effects.OldB299HedronAlignmentReveal), "Hedron Alignment — reveal your hand?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{1}{U}: Scry 1.",
			Purpose: game.Purpose{Scry: 1, Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{1}{U}"),
			Effect:  effects.Do(effects.Scry{N: 1}),
		}},
	})
}
