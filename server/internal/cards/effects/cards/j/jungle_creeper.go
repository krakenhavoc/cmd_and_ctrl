package cardsj

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Jungle Creeper — Creature — Elemental {1}{B}{G}, 3/3:
//
//	"{3}{B}{G}: Return this card from your graveyard to your hand."
//
// An activated ability that works only from the graveyard (CR 113.6b); its
// controller is the card's owner there (CR 108.4), and it returns the card
// unless it has left the graveyard by resolution.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "acc091c0-3891-41b9-b5f6-7537fe9493fe",
		Name:         "Jungle Creeper",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{3}{B}{G}: Return this card from your graveyard to your hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.ManaCost("{3}{B}{G}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisFromYourGraveyardToHand,
		}},
	})
}
