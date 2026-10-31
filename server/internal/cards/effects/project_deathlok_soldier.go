package effects

import "github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

// Project Deathlok Soldier — Artifact Creature — Zombie Soldier {B}, 1/2
// (Marvel Super Heroes, #4455):
//
//	"{2}{B}: Return this card from your graveyard to your hand."
//
// A graveyard-zone activated ability (Deathless Pilot, Blessed Ghoul): the
// row is offered only while the card is in its owner's graveyard, and the
// body does nothing if the card has left since.
//
// No simplification.
func init() {
	Register(Spec{
		OracleID:     "d89aa87a-d59a-41e2-b92f-15c9bc6d3d13",
		Name:         "Project Deathlok Soldier",
		Completeness: CompletenessFull,
		Activated: []ActivatedAbility{{
			Label:   "{2}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    ManaCost("{2}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  returnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
