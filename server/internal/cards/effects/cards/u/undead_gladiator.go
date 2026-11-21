package cardsu

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Undead Gladiator — Creature — Zombie Barbarian {1}{B}{B}, 3/1
// (Dominaria Remastered):
//
//	"{1}{B}, Discard a card: Return this card from your graveyard to your
//	 hand. Activate only during your upkeep.
//	 Cycling {1}{B}"
//
// Eternal Dragon's graveyard return, with a discard in its cost and the
// "only during your upkeep" timing (DuringYourUpkeep, checked as the
// ability is activated). A card that left the graveyard in response stays
// where it is (CR 608.2b). Cycling is the keyword's own ability.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "4da2425d-2118-4116-855d-2ef80693a7b7",
		Name:         "Undead Gladiator",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{
			{
				Label:     "{1}{B}, Discard a card: Return this card from your graveyard to your hand. Activate only during your upkeep.",
				Purpose:   game.Purpose{Answers: game.AnswerValue},
				Cost:      effects.Plus(effects.ManaCost("{1}{B}"), effects.DiscardACard()),
				Zones:     []game.ZoneKind{game.ZoneGraveyard},
				Condition: effects.DuringYourUpkeep(),
				Effect:    effects.ReturnThisCardFromYourGraveyardToYourHand,
			},
			effects.Cycling("{1}{B}"),
		},
	})
}
