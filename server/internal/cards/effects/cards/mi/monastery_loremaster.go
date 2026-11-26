package cardsmi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Monastery Loremaster — Creature — Djinn Wizard {3}{U}, 3/2 (Dragons of Tarkir,
// S82 old batch 331):
//
//	"Megamorph {5}{U} (You may cast this card face down as a 2/2 creature for
//	 {3}. Turn it face up any time for its megamorph cost and put a +1/+1
//	 counter on it.)
//	 When this creature is turned face up, return target noncreature, nonland
//	 card from your graveyard to your hand."
//
// Megamorph is the shared constructor. The face-up trigger takes a required
// target as it goes on the stack (Silumgar Assassin's shape), a noncreature,
// nonland card in your graveyard (Monastery Messenger's filter), and returns it
// if it is still there as the ability resolves (CR 608.2b).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:         "1c8ef715-3c3c-4d04-9f30-941b870b11d3",
		Name:             "Monastery Loremaster",
		Completeness:     effects.CompletenessFull,
		AlternativeCosts: []game.AlternativeCost{effects.Megamorph("{5}{U}")},
		Triggered: []game.TriggeredAbility{
			monasteryLoremasterTurnedFaceUp(),
		},
	})
}

func monasteryLoremasterTurnedFaceUp() game.TriggeredAbility {
	t := effects.WhenThisIsTurnedFaceUp("Monastery Loremaster — return target noncreature, nonland card from your graveyard to your hand",
		effects.B35ReturnChosenGraveyardCardToHand)
	t.Targets = effects.TargetCardInGraveyard("target noncreature, nonland card from your graveyard",
		effects.Noncreature(), effects.Nonland(), effects.YouOwn())
	return t
}
