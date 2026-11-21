package cardsto

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Tunnel Rats — {1}{B} Creature — Rat 2/2 (Teenage Mutant Ninja Turtles,
// #4457):
//
//	"{4}{B}: Return this card from your graveyard to the battlefield
//	 tapped."
//
// Reassembling Skeleton's ability at a different cost: it works only from
// the graveyard, has no timing restriction, and the card enters tapped
// (returnThisFromGraveyardTapped). A card that left the graveyard in
// response stays where it is.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "0a8e5c48-885d-4ea8-9c68-a3e9213d4289",
		Name:         "Tunnel Rats",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{4}{B}: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{4}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisFromGraveyardTapped,
		}},
	})
}
