package effectstest

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// expensiveFrontCheapBack is the Beanstalk Giant // Fertile Footsteps
// shape: a mana value 7 creature whose Adventure is a mana value 3
// sorcery.
func ExpensiveFrontCheapBack(owner uuid.UUID) game.Card {
	c := game.Card{
		InstanceID: uuid.New(), Layout: game.LayoutAdventure, Owner: owner, Controller: owner,
		Faces: []game.Face{
			{Name: "Fixture Beanstalk", TypeLine: "Creature — Giant", ManaCost: "{6}{G}", Colors: []string{"G"},
				Power: 7, Toughness: 7},
			{Name: "Fixture Footsteps", TypeLine: "Sorcery — Adventure", ManaCost: "{2}{G}", Colors: []string{"G"}},
		},
	}
	c.SetFace(0)
	return c
}

// stackCardOf is the card on the stack with this ID.
func StackCardOf(g *game.Game, id uuid.UUID) (game.Card, bool) {
	for _, c := range g.Stack.Cards {
		if c.InstanceID == id {
			return c, true
		}
	}
	return game.Card{}, false
}
