package cardsv

import (
	"github.com/google/uuid"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Vivien's Jaguar — Creature — Cat Spirit {2}{G}, 3/2 (Core Set 2019, S82 OLD batch 198, #5527):
//
//	"Reach
//	 {2}{G}: Return this card from your graveyard to your hand. Activate only
//	 if you control a Vivien planeswalker."
//
// Firewing Phoenix's graveyard return with an activation condition read off
// the battlefield (a planeswalker with the Vivien subtype, as Vivien's
// Crocodile reads it), judged at activation and shown in the view.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "32412a41-2929-4417-90bc-21c5659eeb72",
		Name:            "Vivien's Jaguar",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"reach"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{G}: Return this card from your graveyard to your hand. Activate only if you control a Vivien planeswalker.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Condition: func(g *game.Game, controller, _ uuid.UUID) bool {
				return effects.B42ControlsPlaneswalkerNamed(g, controller, "Vivien")
			},
			Effect: effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
