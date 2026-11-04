package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Bubbling Cauldron — Artifact {2}:
//
//	"{1}, {T}, Sacrifice a creature: You gain 4 life.
//	 {1}, {T}, Sacrifice a creature named Festering Newt: Each opponent
//	 loses 4 life. You gain life equal to the life lost this way."
//
// Two sacrifice outlets on one {T}. The second's cost admits only a
// creature named Festering Newt, and the gain is the life the opponents
// actually lost, summed (extort's drain): an opponent whose life can't
// change gives nothing. A Newt sacrificed to either still triggers its
// own dies ability.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "ca9dde44-fdee-4bc1-a2ea-ef171737b9e4",
		Name:         "Bubbling Cauldron",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{
			{
				Label:   "{1}, {T}, Sacrifice a creature: You gain 4 life.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeACreature()),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return GainLife{Player: item.Controller, Amount: 4}.Apply(NewContext(g, item))
				},
			},
			{
				Label:   "{1}, {T}, Sacrifice a creature named Festering Newt: Each opponent loses 4 life. You gain life equal to the life lost this way.",
				Purpose: game.Purpose{Answers: game.AnswerSacOutlet},
				Cost:    Plus(ManaCost("{1}"), TapCost(), SacrificeN(1, "a creature named Festering Newt", Creature(), jmpB1NamedFesteringNewt)),
				Effect: func(g *game.Game, item *game.StackItem) error {
					return b21DrainEachOpponentAndGainTheTotal(g, item, 4)
				},
			},
		},
	})
}
