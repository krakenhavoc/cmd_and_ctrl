package cardsg

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Gilded Assault Cart — Artifact — Vehicle {1}{R}{R}, 5/1:
//
//	"Trample
//	 Crew 2
//	 Sacrifice two Treasures: Return this card from your graveyard to your hand."
//
// Big Wheel's crew ability, with Blessed Ghoul's graveyard activation: the
// ability is offered only while the card is in its owner's graveyard, and the
// cost is two Treasures sacrificed (any two artifacts with the Treasure
// subtype, tokens or not). The card returns only if it is still in the
// graveyard as the ability resolves (CR 400.7).
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "73da9297-4c7c-4388-820f-1189b5aa0b98",
		Name:            "Gilded Assault Cart",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"trample"},
		Activated: []effects.ActivatedAbility{
			{
				Label:   "Crew 2",
				Cost:    effects.CrewCost(2),
				Purpose: game.Purpose{Answers: game.AnswerAnimate},
				Effect:  effects.CrewEffect("Gilded Assault Cart"),
			},
			{
				Label:   "Sacrifice two Treasures: Return this card from your graveyard to your hand.",
				Purpose: game.Purpose{Answers: game.AnswerValue},
				Cost:    effects.SacrificeN(2, "two Treasures", effects.HasSubtype("Treasure")),
				Zones:   []game.ZoneKind{game.ZoneGraveyard},
				Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
			},
		},
	})
}
