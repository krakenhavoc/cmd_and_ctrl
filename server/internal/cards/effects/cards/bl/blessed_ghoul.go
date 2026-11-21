package cardsbl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Blessed Ghoul — Creature — Zombie Cleric {W/B}, 1/1:
//
//	"Lifelink
//	 {2}{W/B}: Return this card from your graveyard to your hand."
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:        "fe99d90e-4809-4310-a3f7-ceb540511d55",
		Name:            "Blessed Ghoul",
		Completeness:    effects.CompletenessFull,
		PrintedKeywords: []string{"lifelink"},
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{W/B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{W/B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
