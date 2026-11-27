package cardskr

import (
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/cards/effects"
	"github.com/krakenhavoc/cmd_and_ctrl/server/internal/game"
)

// Kyren Archive — Artifact — Book {3} (Mercadian Masques, S82 OLD b682,
// #5619):
//
//	"At the beginning of your upkeep, you may exile the top card of your
//	 library face down.
//	 {5}, Discard your hand, Sacrifice this artifact: Put all cards exiled
//	 with this artifact into their owner's hand."
//
// Asmodeus the Archfiend's blind exile (ExileFaceDownLinkedForEffect): the
// card lands face down so that nobody, its controller included, may look at
// it (CR 406.3), and is linked to this Archive object (CR 607.2a). The upkeep
// trigger asks "you may" before anything is exiled. The ability reads the
// link off the stack item, so the Archive being sacrificed as a cost does not
// lose it; every linked card still in exile goes to its owner's hand.
//
// No simplification.
func init() {
	effects.Register(effects.Spec{
		OracleID:     "f8c7da00-69e1-459d-9b66-6de649c5c511",
		Name:         "Kyren Archive",
		Completeness: effects.CompletenessFull,
		Triggered: []game.TriggeredAbility{
			effects.Optional(effects.AtYourUpkeep("Kyren Archive — exile the top card of your library face down", effects.OldB682ExileTopFaceDownLinked),
				"Exile the top card of your library face down?"),
		},
		Activated: []effects.ActivatedAbility{{
			Label:   "{5}, Discard your hand, Sacrifice this artifact: Put all cards exiled with this artifact into their owner's hand.",
			Purpose: game.Purpose{Answers: game.AnswerValue},
			Cost:    effects.Plus(effects.ManaCost("{5}"), effects.DiscardYourHand(), effects.SacrificeThis()),
			Effect:  effects.B36ReturnCardsExiledWithToOwnersHands,
		}},
	})
}
