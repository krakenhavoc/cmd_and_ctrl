package cardslu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Magus of the Bridge — Creature — Human Wizard {B}{B}{B}, 4/4 (Modern
// Horizons 2):
//
//	"Whenever a nontoken creature is put into your graveyard from the
//	 battlefield, create a 2/2 black Zombie creature token.
//	 When a creature is put into an opponent's graveyard from the battlefield,
//	 exile this creature."
//
// Both are leaves-the-battlefield triggers read off the card's OWNER, since a
// graveyard belongs to the card's owner (CR 404.2), not to whoever controlled
// it. The Magus sees its own death into your graveyard; the second trigger
// exiles only the Magus the ability came from, and does nothing once it has
// left (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "e48951ed-1c32-449a-898e-e5b1c0d0ee73",
		Name:         "Magus of the Bridge",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.On(game.EventLTB, effects.OldB033NontokenCreatureToYourGraveyard,
				"Magus of the Bridge — create a 2/2 black Zombie creature token",
				effects.Do(effects.CreateToken{Template: effects.TokenCard("2/2 black Zombie"), N: 1})),
			effects.On(game.EventLTB, effects.OldB033CreatureToAnOpponentsGraveyard,
				"Magus of the Bridge — exile this creature",
				effects.OldB814ExileThisCreature),
		},
	})
}
