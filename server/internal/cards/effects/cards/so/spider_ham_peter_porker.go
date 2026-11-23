package cardsso

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Spider-Ham, Peter Porker — Legendary Creature — Spider Boar Hero {1}{G}, 2/2
// (Marvel's Spider-Man, #4460):
//
//	"When Spider-Ham enters, create a Food token. (It's an artifact with "{2},
//	 {T}, Sacrifice this token: You gain 3 life.")
//	 Animal May-Ham — Other Spiders, Boars, Bats, Bears, Birds, Cats, Dogs,
//	 Frogs, Jackals, Lizards, Mice, Otters, Rabbits, Raccoons, Rats, Squirrels,
//	 Turtles, and Wolves you control get +1/+1."
//
// Colossus's tribal anthem with a list of types: TribalAnthem(Others,
// YoursOnly) matches a creature that has any of the named types, read through
// the layered view. City Pigeon's Food token on the enters trigger. "Animal
// May-Ham" is an ability word, so it rides only on the stack label.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "00b50215-e832-417f-9c80-2bc3050b6ebb",
		Name:         "Spider-Ham, Peter Porker",
		Completeness: effects.CompletenessFull,
		Static: []game.StaticAbility{
			effects.TribalAnthem(effects.TribeFilter{
				Tribes: []string{"Spider", "Boar", "Bat", "Bear", "Bird", "Cat", "Dog", "Frog", "Jackal",
					"Lizard", "Mouse", "Otter", "Rabbit", "Raccoon", "Rat", "Squirrel", "Turtle", "Wolf"},
				Others:    true,
				YoursOnly: true,
			}, 1, 1),
		},
		Triggered: []game.TriggeredAbility{
			effects.WhenThisEnters("Spider-Ham — create a Food token", effects.Do(effects.CreateToken{Template: effects.FoodToken(), N: 1})),
		},
	})
}
