package cardshu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Ice Cream Kitty — {1}{B/G} Artifact Creature — Food Cat Mutant 1/3
// (Teenage Mutant Ninja Turtles, #4457):
//
//	"{2}, Sacrifice another creature or token: Draw a card. Activate
//	 only as a sorcery.
//	 {2}, {T}, Sacrifice this creature: You gain 3 life."
//
// The first cost takes any other permanent that is a creature, a token or
// both (a Treasure, a Mutant token, a nontoken Bear), never the Kitty
// itself; it is timed as a sorcery (CR 602.5d). The second is
// Gingerbrute's Food ability; as a creature's {T} ability it waits out
// summoning sickness (CR 302.6).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8131dca-ba6e-4ca4-943e-a3d26c75b8b8",
		Name:         "Ice Cream Kitty",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:        "{2}, Sacrifice another creature or token: Draw a card. Activate only as a sorcery.",
				Purpose:      game.Purpose{Draws: 1}, // a sorcery-speed row answers nothing (ADR 0142)
				Cost:         effects.Plus(effects.ManaCost("{2}"), effects.SacrificeAnotherN(1, "another creature or token", effects.Or(effects.Creature(), effects.IsTokenPredicate()))),
				SorcerySpeed: true,
				Effect:       effects.Do(effects.DrawCards{N: 1}),
			},
			{
				Label:   "{2}, {T}, Sacrifice this creature: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.Plus(effects.ManaCost("{2}"), effects.TapCost(), effects.SacrificeThis()),
				Effect:  effects.Do(effects.GainLife{Amount: 3}),
			},
		},
	})
}
