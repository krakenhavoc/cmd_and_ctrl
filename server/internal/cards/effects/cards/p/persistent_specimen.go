package cardsp

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Persistent Specimen — Creature — Skeleton {B}, 1/1 (Innistrad: Crimson Vow):
//
//	"{2}{B}: Return this card from your graveyard to the battlefield tapped."
//
// An ability that functions from the graveyard (CR 113.6): it does nothing if
// the card has left the graveyard by the time it resolves (CR 602.5), and the
// card enters tapped through the same return Reassembling Skeleton uses.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "5f4ad578-f2c4-47f6-8080-8fbc17383f6f",
		Name:         "Persistent Specimen",
		Completeness: effects.CompletenessFull,
		Activated: []effects.ActivatedAbility{{
			Label:   "{2}{B}: Return this card from your graveyard to the battlefield tapped.",
			Purpose: game.Purpose{Answers: game.AnswerMakesBlocker},
			Cost:    effects.ManaCost("{2}{B}"),
			Zones:   []game.ZoneKind{game.ZoneGraveyard},
			Effect:  effects.ReturnThisFromGraveyardTapped,
		}},
	})
}
