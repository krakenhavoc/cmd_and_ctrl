package cardsfl

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"

	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
)

// Fuming Effigy — Creature — Spirit {3}{R}, 4/3 (Strixhaven: School of Mages,
// #5502):
//
//	"Whenever one or more cards leave your graveyard, this creature deals 1
//	 damage to each opponent."
//
// Spirit Mascot's trigger (left_graveyard.go): every route a card leaves
// your graveyard by counts (a cast or a land play from it too), and cards
// leaving together are one batch (OncePerBatch, CR 603.2c). A token ceasing
// to exist is not a card leaving.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "37d48f9b-0286-4c5b-b549-baccb473aea3",
		Name:         "Fuming Effigy",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.OncePerBatch(effects.OnAny(effects.LeftGraveyardEvents, effects.WhenCardsLeaveYourGraveyard,
				"Fuming Effigy — 1 damage to each opponent",
				effects.Do(effects.DamageEachOpponent{N: 1}))),
		},
	})
}
