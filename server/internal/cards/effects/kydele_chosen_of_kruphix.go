package effects

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kydele, Chosen of Kruphix — Legendary Creature — Human Wizard
// {2}{G}{U}, 2/3 (Commander 2016, #5537):
//
//	"{T}: Add {C} for each card you've drawn this turn.
//	 Partner (You can have two commanders if both have partner.)"
//
// Immortus's count-driven mana ability with colorless output and no
// spending restriction: the controller's tally of cards drawn this turn,
// read as the ability is activated (a card drawn and then discarded
// still counts). With no draw this turn it adds nothing. Partner is a
// deck-building rule (CR 702.124) that internal/deck reads off the
// oracle text.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "5fca0ce6-e6ba-4319-b3c4-b4d48901e9b4",
		Name:         "Kydele, Chosen of Kruphix",
		Completeness: CompletenessFull,
		ManaAbilities: []ManaAbility{{
			Cost:           ManaAbilityCost{Tap: true},
			ProducedAmount: kydeleProduced,
			Label:          "Add {C} for each card you've drawn this turn",
		}},
	})
}

func kydeleProduced(g *game.Game, controller, _ uuid.UUID) game.ManaAmount {
	return game.ManaAmount{Options: []string{"C"}, N: g.TurnTallyFor(controller).CardsDrawn}
}
