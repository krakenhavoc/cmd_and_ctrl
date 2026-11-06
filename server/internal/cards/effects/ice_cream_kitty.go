package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

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
	Register(Spec{
		OracleID:     "f8131dca-ba6e-4ca4-943e-a3d26c75b8b8",
		Name:         "Ice Cream Kitty",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:        "{2}, Sacrifice another creature or token: Draw a card. Activate only as a sorcery.",
				Purpose:      game.Purpose{Draws: 1}, // a sorcery-speed row answers nothing (ADR 0142)
				Cost:         Plus(ManaCost("{2}"), SacrificeAnotherN(1, "another creature or token", Or(Creature(), IsTokenPredicate()))),
				SorcerySpeed: true,
				Effect:       Do(DrawCards{N: 1}),
			},
			{
				Label:   "{2}, {T}, Sacrifice this creature: You gain 3 life.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    Plus(ManaCost("{2}"), TapCost(), SacrificeThis()),
				Effect:  Do(GainLife{Amount: 3}),
			},
		},
	})
}
