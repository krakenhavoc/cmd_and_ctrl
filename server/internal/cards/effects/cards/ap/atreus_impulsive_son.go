package cardsap

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Atreus, Impulsive Son — Legendary Creature — God Archer {1}{U}{R}, 2/4
// (Secret Lair Drop):
//
//	"Reach
//	 {3}, {T}: Draw a card for each experience counter you have, then discard a
//	 card. Atreus deals 2 damage to each opponent.
//	 Partner—Father & son (You can have two commanders if both have this
//	 ability.)"
//
// The draw counts the experience counters as the ability resolves (Kratos's
// reading), then the discard is the loot's: a prompt opened after the draws, so
// a drawn card is a legal discard (CR 608.2c), and a player with no experience
// counters still discards. The damage is the Atreus's own to each opponent
// (DamageEachOpponent), dealt in the same resolution. Partner—Father & son is a
// deck-construction rule read off the oracle text.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "f6a4298c-3d9d-43e1-b32c-b168c02ad503",
		Name:            "Atreus, Impulsive Son",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}, {T}: Draw a card for each experience counter you have, then discard a card. Atreus deals 2 damage to each opponent.",
			Purpose: game.Purpose{Draws: 1, Discards: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{3}"), effects.TapCost()),
			Effect: func(g *game.Game, item *game.StackItem) error {
				// The loot's discard waits on a prompt, so the damage is queued
				// first: it neither reads nor changes anything the loot does.
				if err := effects.DamageToEachOpponent(g, item, 2); err != nil {
					return err
				}
				return effects.DrawThenDiscard(g, item.Controller, item.SourceCardID, effects.ExperienceCounters(g, item.Controller), 1, "")
			},
		}},
	})
}
