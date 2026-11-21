package cardss

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Sanitarium Skeleton — Creature — Skeleton {B}, 1/2 (Jumpstart, #4892):
//
//	"{2}{B}: Return this card from your graveyard to your hand."
//
// Activated only from the graveyard (CR 113.6). If the card has left the
// graveyard by the time the ability resolves, nothing happens.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "d09b0648-e324-4a16-b78c-d1849177c04f",
		Name:         "Sanitarium Skeleton",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{2}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisCardFromYourGraveyardToYourHand,
		}},
	})
}
