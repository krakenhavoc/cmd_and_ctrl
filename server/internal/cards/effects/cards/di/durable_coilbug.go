package cardsdi

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Durable Coilbug — Creature — Insect {1}{B}, 2/2 (Foundations Jumpstart):
//
//	"{4}{B}: Return this card from your graveyard to your hand."
//
// Abzan Devotee's graveyard ability: it works only from the graveyard, and
// does nothing if the card has left it by resolution (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "a421582a-f673-4f43-beb5-28ee785da11e",
		Name:         "Durable Coilbug",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
