package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// gapB31EternalizeCreature is Eternalize for a card whose reminder text says
// the token "loses all other card types" (Lazotep Archway): the ordinary
// eternalize exception, then a type line of just Token Creature — Zombie.
func gapB31EternalizeCreature(cost string) ActivatedAbility {
	ab := Eternalize(cost)
	ab.Label = "Eternalize " + cost + " (" + cost + ", Exile this card from your graveyard: " +
		"Create a token that's a copy of it, except it's a 4/4 black Zombie creature and loses all " +
		"other card types. Activate only as a sorcery.)"
	ab.Effect = func(g *game.Game, item *game.StackItem) error {
		return zombieTokenCopyOfSource(g, item, gapB31CreatureOnly)
	}
	return ab
}

// gapB31CreatureOnly is the exception: eternalize's, then "loses all other
// card types".
func gapB31CreatureOnly(t *game.Card) {
	eternalizeException(t)
	t.TypeLine = "Token Creature — Zombie"
}
