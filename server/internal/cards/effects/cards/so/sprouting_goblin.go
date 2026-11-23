package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sprouting Goblin — Creature — Goblin Druid {1}{R}, 2/2 (Foundations
// Commander):
//
//	"Kicker {G} (You may pay an additional {G} as you cast this spell.)
//	 When this creature enters, if it was kicked, search your library for
//	 a land card with a basic land type, reveal it, put it into your
//	 hand, then shuffle.
//	 {R}, {T}, Sacrifice a land: Draw a card."
//
// Ana Battlemage's kicked enters trigger over Boseiju's "land card with a
// basic land type" predicate (a shockland or a Triome qualifies, a Wastes
// does not) and Abzan Monument's reveal-to-hand search. The draw pays a
// land sacrifice as a cost.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:      "b0871e12-b156-4fa6-b01c-acaae222b3a5",
		Name:          "Sprouting Goblin",
		Completeness:  effects.CompletenessFull,
		Purpose:       game.Purpose{Tutors: 1},
		OptionalCosts: []game.AdditionalCost{effects.Kicker("{G}")},
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventETB, effects.AllOf(effects.Self, effects.ThisKickedWith("{G}")),
				"Sprouting Goblin — kicked: search for a land card with a basic land type",
				func(g *game.Game, item *game.StackItem) error {
					return effects.SearchLibrary{
						Player:    item.Controller,
						Source:    item.SourceCardID,
						Predicate: effects.HasABasicLandType,
						Dest:      game.ZoneHand,
						Limit:     1,
						Reveal:    true,
						Shuffle:   true,
						Reason:    "Sprouting Goblin — a land card with a basic land type",
					}.Apply(effects.NewContext(g, item))
				}),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{R}, {T}, Sacrifice a land: Draw a card.",
			Purpose: game.Purpose{Draws: 1, Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{R}"), effects.TapCost(), effects.B08SacrificeALand()),
			Effect:  effects.Do(effects.DrawCards{N: 1}),
		}},
	})
}
